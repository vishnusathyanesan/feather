package dm

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

var (
	ErrSelfDM    = errors.New("cannot create DM with yourself")
	ErrDMTooFew  = errors.New("group DM requires at least 2 other users")
	ErrDMTooMany = errors.New("group DM max 8 members")
)

type BroadcastFunc func(channelID uuid.UUID, event model.WebSocketEvent)

type SubscribeFn func(userID uuid.UUID, channelID uuid.UUID)

type Service struct {
	repo        *Repository
	broadcast   BroadcastFunc
	subscribeFn SubscribeFn
}

func NewService(repo *Repository, broadcast BroadcastFunc, subscribeFn SubscribeFn) *Service {
	return &Service{repo: repo, broadcast: broadcast, subscribeFn: subscribeFn}
}

// GetOrCreateDM finds existing 1:1 DM or creates a new one.
func (s *Service) GetOrCreateDM(ctx context.Context, userID, otherUserID uuid.UUID) (*model.Channel, error) {
	if userID == otherUserID {
		return nil, ErrSelfDM
	}

	existing, err := s.repo.FindExistingDM(ctx, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		members, err := s.repo.GetDMMembers(ctx, existing.ID)
		if err == nil {
			existing.Members = members
		}
		return existing, nil
	}

	now := time.Now()
	ch := &model.Channel{
		ID:        uuid.New(),
		Type:      model.ChannelDM,
		CreatorID: &userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	memberIDs := []uuid.UUID{userID, otherUserID}
	if err := s.repo.CreateDMChannel(ctx, ch, memberIDs); err != nil {
		return nil, err
	}

	members, _ := s.repo.GetDMMembers(ctx, ch.ID)
	ch.Members = members
	ch.MemberCount = len(memberIDs)

	// Subscribe both users to the new channel for live updates
	if s.subscribeFn != nil {
		for _, id := range memberIDs {
			s.subscribeFn(id, ch.ID)
		}
	}

	// Broadcast dm.created event to both users
	if s.broadcast != nil {
		payload, _ := json.Marshal(ch)
		event := model.WebSocketEvent{
			Type:      model.EventDMCreated,
			ChannelID: ch.ID.String(),
			Payload:   payload,
		}
		s.broadcast(ch.ID, event)
	}

	return ch, nil
}

// CreateGroupDM creates a new group DM.
func (s *Service) CreateGroupDM(ctx context.Context, creatorID uuid.UUID, userIDs []uuid.UUID) (*model.Channel, error) {
	if len(userIDs) < 2 {
		return nil, ErrDMTooFew
	}
	if len(userIDs) > 8 {
		return nil, ErrDMTooMany
	}

	now := time.Now()
	ch := &model.Channel{
		ID:        uuid.New(),
		Type:      model.ChannelGroupDM,
		CreatorID: &creatorID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Include creator in members and deduplicate
	allMembers := append([]uuid.UUID{creatorID}, userIDs...)
	seen := make(map[uuid.UUID]bool)
	var unique []uuid.UUID
	for _, id := range allMembers {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	// Check for existing group DM with the same members
	existing, err := s.repo.FindExistingGroupDM(ctx, unique)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		members, err := s.repo.GetDMMembers(ctx, existing.ID)
		if err == nil {
			existing.Members = members
		}
		return existing, nil
	}

	if err := s.repo.CreateDMChannel(ctx, ch, unique); err != nil {
		return nil, err
	}

	members, _ := s.repo.GetDMMembers(ctx, ch.ID)
	ch.Members = members
	ch.MemberCount = len(unique)

	if s.subscribeFn != nil {
		for _, id := range unique {
			s.subscribeFn(id, ch.ID)
		}
	}

	if s.broadcast != nil {
		payload, _ := json.Marshal(ch)
		event := model.WebSocketEvent{
			Type:      model.EventDMCreated,
			ChannelID: ch.ID.String(),
			Payload:   payload,
		}
		s.broadcast(ch.ID, event)
	}

	return ch, nil
}

// ListDMs returns all DM conversations for a user.
func (s *Service) ListDMs(ctx context.Context, userID uuid.UUID) ([]model.Channel, error) {
	dms, err := s.repo.ListDMs(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Populate members for each DM
	for i := range dms {
		members, err := s.repo.GetDMMembers(ctx, dms[i].ID)
		if err == nil {
			dms[i].Members = members
		}
	}

	return dms, nil
}
