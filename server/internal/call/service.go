package call

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

var (
	ErrCallNotFound     = errors.New("call not found")
	ErrCallAlreadyActive = errors.New("there is already an active call in this channel")
	ErrNotCallParticipant = errors.New("not a call participant")
	ErrCallNotRinging   = errors.New("call is not in ringing state")
)

const ringingTimeout = 30 * time.Second

type BroadcastFunc func(channelID uuid.UUID, event model.WebSocketEvent)
type SendToUserFunc func(userID uuid.UUID, data []byte)

// ChannelMemberChecker checks if a user is a member of a channel.
type ChannelMemberChecker interface {
	IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
}

type Service struct {
	repo          *Repository
	broadcast     BroadcastFunc
	sendToUser    SendToUserFunc
	memberChecker ChannelMemberChecker
	timers        map[uuid.UUID]*time.Timer
	timersMu      sync.Mutex
}

func NewService(repo *Repository, broadcast BroadcastFunc, sendToUser SendToUserFunc) *Service {
	return &Service{
		repo:       repo,
		broadcast:  broadcast,
		sendToUser: sendToUser,
		timers:     make(map[uuid.UUID]*time.Timer),
	}
}

// SetMemberChecker sets the channel membership checker for authorization.
func (s *Service) SetMemberChecker(mc ChannelMemberChecker) {
	s.memberChecker = mc
}

func (s *Service) Initiate(ctx context.Context, req model.InitiateCallRequest, initiatorID uuid.UUID) (*model.Call, error) {
	// Check for existing active call
	existing, err := s.repo.GetActiveCallForChannel(ctx, req.ChannelID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCallAlreadyActive
	}

	c := &model.Call{
		ID:          uuid.New(),
		ChannelID:   req.ChannelID,
		InitiatorID: initiatorID,
		CallType:    req.CallType,
		Status:      model.CallStatusRinging,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}

	// Add initiator as participant
	_ = s.repo.AddParticipant(ctx, c.ID, initiatorID)

	// Broadcast ringing event
	s.broadcastCallEvent(model.EventCallRinging, c)

	// Start ringing timeout
	s.startRingingTimeout(c.ID, c.ChannelID)

	return c, nil
}

func (s *Service) Accept(ctx context.Context, callID, userID uuid.UUID) (*model.Call, error) {
	c, err := s.repo.GetByID(ctx, callID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCallNotFound
	}
	if c.Status != model.CallStatusRinging {
		return nil, ErrCallNotRinging
	}

	s.cancelTimer(callID)

	if err := s.repo.SetStarted(ctx, callID); err != nil {
		return nil, err
	}
	_ = s.repo.AddParticipant(ctx, callID, userID)

	c.Status = model.CallStatusInProgress
	c.AcceptedBy = userID
	s.broadcastCallEvent(model.EventCallAccepted, c)

	return c, nil
}

func (s *Service) Decline(ctx context.Context, callID, userID uuid.UUID) error {
	c, err := s.repo.GetByID(ctx, callID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCallNotFound
	}

	// Verify the user is a member of the call's channel
	if s.memberChecker != nil {
		isMember, err := s.memberChecker.IsMember(ctx, c.ChannelID, userID)
		if err != nil {
			return err
		}
		if !isMember {
			return ErrNotCallParticipant
		}
	}

	s.cancelTimer(callID)

	if err := s.repo.SetEnded(ctx, callID, model.CallStatusDeclined); err != nil {
		return err
	}

	c.Status = model.CallStatusDeclined
	s.broadcastCallEvent(model.EventCallDeclined, c)
	return nil
}

func (s *Service) Hangup(ctx context.Context, callID, userID uuid.UUID) error {
	c, err := s.repo.GetByID(ctx, callID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCallNotFound
	}

	s.cancelTimer(callID)
	_ = s.repo.RemoveParticipant(ctx, callID, userID)

	// Only end the call if no active participants remain
	remaining, err := s.repo.CountActiveParticipants(ctx, callID)
	if err != nil {
		return err
	}

	if remaining == 0 {
		if err := s.repo.SetEnded(ctx, callID, model.CallStatusEnded); err != nil {
			return err
		}
		c.Status = model.CallStatusEnded
		s.broadcastCallEvent(model.EventCallEnded, c)
	}

	return nil
}

func (s *Service) GetActiveCall(ctx context.Context, channelID uuid.UUID) (*model.Call, error) {
	return s.repo.GetActiveCallForChannel(ctx, channelID)
}

func (s *Service) ListByChannel(ctx context.Context, channelID uuid.UUID) ([]model.Call, error) {
	return s.repo.ListByChannel(ctx, channelID, 20)
}

// RecoverStaleCalls marks any ringing calls older than the ringing timeout as missed.
// Should be called on server startup to clean up calls left in ringing state from a prior crash.
func (s *Service) RecoverStaleCalls(ctx context.Context) {
	cutoff := time.Now().Add(-ringingTimeout)
	count, err := s.repo.ExpireStaleRingingCalls(ctx, cutoff)
	if err != nil {
		slog.Error("failed to recover stale ringing calls", "error", err)
		return
	}
	if count > 0 {
		slog.Info("recovered stale ringing calls", "count", count)
	}
}

func (s *Service) broadcastCallEvent(eventType model.EventType, c *model.Call) {
	if s.broadcast == nil {
		return
	}
	payload, _ := json.Marshal(c)
	event := model.WebSocketEvent{
		Type:      eventType,
		ChannelID: c.ChannelID.String(),
		Payload:   payload,
	}
	s.broadcast(c.ChannelID, event)
}

func (s *Service) startRingingTimeout(callID, channelID uuid.UUID) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()

	timer := time.AfterFunc(ringingTimeout, func() {
		ctx := context.Background()
		c, err := s.repo.GetByID(ctx, callID)
		if err != nil || c == nil {
			return
		}
		if c.Status != model.CallStatusRinging {
			return
		}

		if err := s.repo.SetEnded(ctx, callID, model.CallStatusMissed); err != nil {
			slog.Error("failed to set call as missed", "error", err, "call_id", callID)
			return
		}

		c.Status = model.CallStatusMissed
		s.broadcastCallEvent(model.EventCallMissed, c)
	})

	s.timers[callID] = timer
}

func (s *Service) cancelTimer(callID uuid.UUID) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()

	if timer, ok := s.timers[callID]; ok {
		timer.Stop()
		delete(s.timers, callID)
	}
}
