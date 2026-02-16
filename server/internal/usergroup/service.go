package usergroup

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

var (
	ErrGroupNotFound = errors.New("user group not found")
	ErrGroupNameTaken = errors.New("group name already taken")
	ErrForbidden     = errors.New("forbidden")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req model.CreateUserGroupRequest, creatorID uuid.UUID) (*model.UserGroup, error) {
	existing, err := s.repo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrGroupNameTaken
	}

	now := time.Now()
	g := &model.UserGroup{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatorID:   creatorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, g); err != nil {
		return nil, err
	}

	// Add creator as first member
	_ = s.repo.AddMember(ctx, g.ID, creatorID)

	return g, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*model.UserGroup, error) {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrGroupNotFound
	}

	members, _ := s.repo.GetMembers(ctx, id)
	g.Members = members
	return g, nil
}

func (s *Service) List(ctx context.Context, search string) ([]model.UserGroup, error) {
	return s.repo.List(ctx, search)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req model.UpdateUserGroupRequest, userID uuid.UUID, userRole string) (*model.UserGroup, error) {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrGroupNotFound
	}

	if g.CreatorID != userID && userRole != string(model.RoleAdmin) {
		return nil, ErrForbidden
	}

	if req.Name != nil {
		existing, err := s.repo.GetByName(ctx, *req.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrGroupNameTaken
		}
		g.Name = *req.Name
	}
	if req.Description != nil {
		g.Description = *req.Description
	}

	if err := s.repo.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string) error {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if g == nil {
		return ErrGroupNotFound
	}

	if g.CreatorID != userID && userRole != string(model.RoleAdmin) {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) AddMember(ctx context.Context, groupID, userID uuid.UUID) error {
	g, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if g == nil {
		return ErrGroupNotFound
	}
	return s.repo.AddMember(ctx, groupID, userID)
}

func (s *Service) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return s.repo.RemoveMember(ctx, groupID, userID)
}
