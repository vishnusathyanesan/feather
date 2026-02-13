package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrNameTaken       = errors.New("channel name already taken")
	ErrNotMember       = errors.New("not a channel member")
	ErrForbidden       = errors.New("forbidden")
	ErrAlreadyMember   = errors.New("already a member")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req model.CreateChannelRequest, userID uuid.UUID, userRole string) (*model.Channel, error) {
	if req.Type == model.ChannelSystem && userRole != string(model.RoleAdmin) {
		return nil, ErrForbidden
	}

	existing, err := s.repo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrNameTaken
	}

	now := time.Now()
	ch := &model.Channel{
		ID:          uuid.New(),
		Name:        req.Name,
		Topic:       req.Topic,
		Description: req.Description,
		Type:        req.Type,
		IsReadonly:  req.IsReadonly,
		CreatorID:   &userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, ch); err != nil {
		return nil, err
	}

	// Auto-add creator as channel admin
	if err := s.repo.AddMember(ctx, ch.ID, userID, "admin"); err != nil {
		return nil, err
	}

	ch.MemberCount = 1
	return ch, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Channel, error) {
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	// Private channels require membership
	if ch.Type == model.ChannelPrivate {
		isMember, err := s.repo.IsMember(ctx, id, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrForbidden
		}
	}

	return ch, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]model.Channel, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req model.UpdateChannelRequest, userID uuid.UUID, userRole string) (*model.Channel, error) {
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	// Only creator or admin can update
	isCreator := ch.CreatorID != nil && *ch.CreatorID == userID
	if !isCreator && userRole != string(model.RoleAdmin) {
		return nil, ErrForbidden
	}

	if req.Name != nil {
		existing, err := s.repo.GetByName(ctx, *req.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrNameTaken
		}
		ch.Name = *req.Name
	}
	if req.Topic != nil {
		ch.Topic = *req.Topic
	}
	if req.Description != nil {
		ch.Description = *req.Description
	}
	if req.IsReadonly != nil {
		ch.IsReadonly = *req.IsReadonly
	}

	if err := s.repo.Update(ctx, ch); err != nil {
		return nil, err
	}

	return ch, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userRole string) error {
	if userRole != string(model.RoleAdmin) {
		return ErrForbidden
	}

	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if ch == nil {
		return ErrChannelNotFound
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) Join(ctx context.Context, channelID, userID uuid.UUID) error {
	ch, err := s.repo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if ch == nil {
		return ErrChannelNotFound
	}

	if ch.Type == model.ChannelPrivate {
		return ErrForbidden
	}

	return s.repo.AddMember(ctx, channelID, userID, "member")
}

func (s *Service) Leave(ctx context.Context, channelID, userID uuid.UUID) error {
	return s.repo.RemoveMember(ctx, channelID, userID)
}

func (s *Service) InviteMember(ctx context.Context, channelID, inviterID, inviteeID uuid.UUID, inviterRole string) error {
	ch, err := s.repo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if ch == nil {
		return ErrChannelNotFound
	}

	// Only channel members (or admins) can invite
	isMember, err := s.repo.IsMember(ctx, channelID, inviterID)
	if err != nil {
		return err
	}
	if !isMember && inviterRole != string(model.RoleAdmin) {
		return ErrForbidden
	}

	return s.repo.AddMember(ctx, channelID, inviteeID, "member")
}

func (s *Service) GetMembers(ctx context.Context, channelID, userID uuid.UUID) ([]model.ChannelMember, error) {
	ch, err := s.repo.GetByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	if ch.Type == model.ChannelPrivate {
		isMember, err := s.repo.IsMember(ctx, channelID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrForbidden
		}
	}

	return s.repo.GetMembers(ctx, channelID)
}

func (s *Service) MarkRead(ctx context.Context, channelID, userID uuid.UUID) error {
	return s.repo.UpdateLastRead(ctx, channelID, userID)
}

func (s *Service) IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return s.repo.IsMember(ctx, channelID, userID)
}

func (s *Service) SeedDefaultChannel(ctx context.Context) (*model.Channel, error) {
	existing, err := s.repo.GetByName(ctx, "general")
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	now := time.Now()
	ch := &model.Channel{
		ID:          uuid.New(),
		Name:        "general",
		Topic:       "General discussion",
		Description: "Default channel for everyone",
		Type:        model.ChannelPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, ch); err != nil {
		return nil, fmt.Errorf("seed general channel: %w", err)
	}

	return ch, nil
}

func (s *Service) AutoJoinUser(ctx context.Context, userID uuid.UUID) error {
	general, err := s.repo.GetByName(ctx, "general")
	if err != nil {
		return err
	}
	if general == nil {
		return nil
	}
	return s.repo.AddMember(ctx, general.ID, userID, "member")
}
