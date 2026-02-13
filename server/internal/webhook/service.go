package webhook

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

var (
	ErrWebhookNotFound = errors.New("webhook not found")
	ErrInvalidToken    = errors.New("invalid webhook token")
)

type MessageCreator interface {
	CreateAlertMessage(ctx context.Context, channelID, botUserID uuid.UUID, content string, severity string, metadata json.RawMessage) (*model.Message, error)
}

type Service struct {
	repo      *Repository
	botUserID uuid.UUID
	msgCreate MessageCreator
}

func NewService(repo *Repository, botUserID uuid.UUID, msgCreate MessageCreator) *Service {
	return &Service{repo: repo, botUserID: botUserID, msgCreate: msgCreate}
}

func (s *Service) Create(ctx context.Context, req model.CreateWebhookRequest, creatorID uuid.UUID) (*model.Webhook, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	wh := &model.Webhook{
		ID:        uuid.New(),
		ChannelID: req.ChannelID,
		Name:      req.Name,
		Token:     token,
		CreatorID: creatorID,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, wh); err != nil {
		return nil, err
	}

	return wh, nil
}

func (s *Service) List(ctx context.Context, creatorID uuid.UUID) ([]model.Webhook, error) {
	return s.repo.List(ctx, creatorID)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string) error {
	wh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if wh == nil {
		return ErrWebhookNotFound
	}
	if wh.CreatorID != userID && userRole != string(model.RoleAdmin) {
		return fmt.Errorf("forbidden")
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) HandleIncoming(ctx context.Context, token string, payload model.WebhookPayload) error {
	wh, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return err
	}
	if wh == nil {
		return ErrInvalidToken
	}

	content := fmt.Sprintf("**%s**\n\n%s", payload.Title, payload.Message)

	if s.msgCreate != nil {
		_, err = s.msgCreate.CreateAlertMessage(ctx, wh.ChannelID, s.botUserID, content, payload.Severity, payload.Metadata)
		return err
	}

	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
