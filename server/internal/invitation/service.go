package invitation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

var (
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationExpired  = errors.New("invitation has expired")
	ErrInvitationMaxUsed  = errors.New("invitation has reached max uses")
	ErrForbidden          = errors.New("forbidden")
)

type AutoJoiner interface {
	AutoJoinUser(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	repo       *Repository
	autoJoiner AutoJoiner
	appURL     string
}

func NewService(repo *Repository, autoJoiner AutoJoiner, appURL string) *Service {
	return &Service{repo: repo, autoJoiner: autoJoiner, appURL: appURL}
}

func (s *Service) Create(ctx context.Context, req model.CreateInvitationRequest, inviterID uuid.UUID) (*model.WorkspaceInvitation, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	inv := &model.WorkspaceInvitation{
		ID:        uuid.New(),
		InviterID: inviterID,
		Email:     req.Email,
		Token:     token,
		ExpiresAt: now.Add(time.Duration(req.TTLDays) * 24 * time.Hour),
		MaxUses:   req.MaxUses,
		CreatedAt: now,
	}

	if err := s.repo.Create(ctx, inv); err != nil {
		return nil, err
	}

	inv.InviteURL = s.buildInviteURL(token)
	return inv, nil
}

func (s *Service) Validate(ctx context.Context, token string) (*model.WorkspaceInvitation, error) {
	inv, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrInvitationNotFound
	}

	if time.Now().After(inv.ExpiresAt) {
		return nil, ErrInvitationExpired
	}
	if inv.UseCount >= inv.MaxUses {
		return nil, ErrInvitationMaxUsed
	}

	inv.InviteURL = s.buildInviteURL(token)
	return inv, nil
}

func (s *Service) Accept(ctx context.Context, token string, userID uuid.UUID) error {
	inv, err := s.Validate(ctx, token)
	if err != nil {
		return err
	}

	if err := s.repo.IncrementUseCount(ctx, inv.ID, userID); err != nil {
		return err
	}

	// Auto-join user to default channels
	if s.autoJoiner != nil {
		_ = s.autoJoiner.AutoJoinUser(ctx, userID)
	}

	return nil
}

func (s *Service) List(ctx context.Context, inviterID uuid.UUID) ([]model.WorkspaceInvitation, error) {
	invitations, err := s.repo.ListByInviter(ctx, inviterID)
	if err != nil {
		return nil, err
	}
	for i := range invitations {
		invitations[i].InviteURL = s.buildInviteURL(invitations[i].Token)
	}
	return invitations, nil
}

func (s *Service) Revoke(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string) error {
	inv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return ErrInvitationNotFound
	}

	if inv.InviterID != userID && userRole != string(model.RoleAdmin) {
		return ErrForbidden
	}

	return s.repo.Revoke(ctx, id)
}

func (s *Service) buildInviteURL(token string) string {
	if s.appURL != "" {
		return s.appURL + "/invite/" + token
	}
	return "/invite/" + token
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
