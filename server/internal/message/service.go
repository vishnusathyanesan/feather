package message

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrForbidden       = errors.New("forbidden")
	ErrEditExpired     = errors.New("edit window expired (24h)")
	ErrReadonly        = errors.New("channel is read-only")
)

type ChannelChecker interface {
	IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
}

type BroadcastFunc func(channelID uuid.UUID, event model.WebSocketEvent)

type Service struct {
	repo      *Repository
	channels  ChannelChecker
	broadcast BroadcastFunc
}

func NewService(repo *Repository, channels ChannelChecker, broadcast BroadcastFunc) *Service {
	return &Service{
		repo:      repo,
		channels:  channels,
		broadcast: broadcast,
	}
}

func (s *Service) Create(ctx context.Context, channelID uuid.UUID, req model.CreateMessageRequest, userID uuid.UUID, userRole string) (*model.Message, error) {
	isMember, err := s.channels.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrForbidden
	}

	msg := &model.Message{
		ID:        uuid.New(),
		ChannelID: channelID,
		UserID:    userID,
		ParentID:  req.ParentID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, err
	}

	// Fetch full message with user data
	full, err := s.repo.GetByID(ctx, msg.ID)
	if err != nil {
		return msg, nil // Return basic message if fetch fails
	}

	// Fetch reactions
	reactions, err := s.repo.GetReactions(ctx, full.ID)
	if err == nil {
		full.Reactions = reactions
	}

	if s.broadcast != nil {
		s.broadcastMessage(model.EventMessageNew, full)
	}

	return full, nil
}

func (s *Service) List(ctx context.Context, params model.MessageListParams, userID uuid.UUID) ([]model.Message, error) {
	isMember, err := s.channels.IsMember(ctx, params.ChannelID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrForbidden
	}

	messages, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	// Batch fetch reactions
	if len(messages) > 0 {
		ids := make([]uuid.UUID, len(messages))
		for i, m := range messages {
			ids[i] = m.ID
		}
		reactions, err := s.repo.GetReactionsForMessages(ctx, ids)
		if err == nil {
			for i := range messages {
				if r, ok := reactions[messages[i].ID]; ok {
					messages[i].Reactions = r
				}
			}
		}
	}

	return messages, nil
}

func (s *Service) GetThread(ctx context.Context, parentID uuid.UUID, userID uuid.UUID) ([]model.Message, error) {
	parent, err := s.repo.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, ErrMessageNotFound
	}

	isMember, err := s.channels.IsMember(ctx, parent.ChannelID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrForbidden
	}

	replies, err := s.repo.GetThread(ctx, parentID)
	if err != nil {
		return nil, err
	}

	// Include parent as first message
	result := append([]model.Message{*parent}, replies...)
	return result, nil
}

func (s *Service) Update(ctx context.Context, channelID, messageID uuid.UUID, req model.UpdateMessageRequest, userID uuid.UUID, userRole string) (*model.Message, error) {
	msg, err := s.repo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, ErrMessageNotFound
	}

	// Only author can edit, within 24h
	if msg.UserID != userID {
		return nil, ErrForbidden
	}
	if time.Since(msg.CreatedAt) > 24*time.Hour {
		return nil, ErrEditExpired
	}

	if err := s.repo.Update(ctx, messageID, req.Content); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	if s.broadcast != nil {
		s.broadcastMessage(model.EventMessageUpdated, updated)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, channelID, messageID uuid.UUID, userID uuid.UUID, userRole string) error {
	msg, err := s.repo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg == nil {
		return ErrMessageNotFound
	}

	// Author or admin can delete
	if msg.UserID != userID && userRole != string(model.RoleAdmin) {
		return ErrForbidden
	}

	if err := s.repo.SoftDelete(ctx, messageID); err != nil {
		return err
	}

	if s.broadcast != nil {
		s.broadcastMessage(model.EventMessageDeleted, msg)
	}

	return nil
}

func (s *Service) broadcastMessage(eventType model.EventType, msg *model.Message) {
	payload, _ := json.Marshal(msg)
	event := model.WebSocketEvent{
		Type:      eventType,
		ChannelID: msg.ChannelID.String(),
		Payload:   payload,
	}
	s.broadcast(msg.ChannelID, event)
}
