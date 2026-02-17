package mention

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

type BroadcastFunc func(channelID uuid.UUID, event model.WebSocketEvent)

type Service struct {
	repo      *Repository
	broadcast BroadcastFunc
}

func NewService(repo *Repository, broadcast BroadcastFunc) *Service {
	return &Service{repo: repo, broadcast: broadcast}
}

// ProcessMentions parses a message for @mentions, resolves them, creates records, and notifies.
func (s *Service) ProcessMentions(ctx context.Context, msg *model.Message) {
	parsed := ParseMentions(msg.Content)
	if len(parsed) == 0 {
		return
	}

	for _, p := range parsed {
		switch p.Type {
		case "channel", "here":
			s.handleChannelMention(ctx, msg, p)
		case "user_or_group":
			s.handleUserOrGroupMention(ctx, msg, p)
		}
	}
}

func (s *Service) handleUserOrGroupMention(ctx context.Context, msg *model.Message, p ParsedMention) {
	// Try user first
	user, err := s.repo.GetUserByName(ctx, p.Name)
	if err != nil {
		slog.Error("mention: failed to resolve user", "name", p.Name, "error", err)
		return
	}
	if user != nil {
		mn := &model.Mention{
			ID:              uuid.New(),
			MessageID:       msg.ID,
			ChannelID:       msg.ChannelID,
			MentionedUserID: &user.ID,
			MentionType:     "user",
			CreatedAt:       time.Now(),
		}
		if err := s.repo.Create(ctx, mn); err != nil {
			slog.Error("mention: failed to create", "error", err)
			return
		}
		s.notifyMention(mn, msg)
		return
	}

	// Try group
	group, err := s.repo.GetGroupByName(ctx, p.Name)
	if err != nil {
		slog.Error("mention: failed to resolve group", "name", p.Name, "error", err)
		return
	}
	if group != nil {
		memberIDs, err := s.repo.GetGroupMemberIDs(ctx, group.ID)
		if err != nil {
			slog.Error("mention: failed to get group members", "group", p.Name, "error", err)
			return
		}
		for _, memberID := range memberIDs {
			if memberID == msg.UserID {
				continue // Don't notify the sender
			}
			uid := memberID
			mn := &model.Mention{
				ID:               uuid.New(),
				MessageID:        msg.ID,
				ChannelID:        msg.ChannelID,
				MentionedUserID:  &uid,
				MentionedGroupID: &group.ID,
				MentionType:      "group",
				CreatedAt:        time.Now(),
			}
			if err := s.repo.Create(ctx, mn); err != nil {
				slog.Error("mention: failed to create group mention", "error", err)
				continue
			}
			s.notifyMention(mn, msg)
		}
	}
}

func (s *Service) handleChannelMention(ctx context.Context, msg *model.Message, p ParsedMention) {
	memberIDs, err := s.repo.GetChannelMemberIDs(ctx, msg.ChannelID)
	if err != nil {
		slog.Error("mention: failed to get channel members", "error", err)
		return
	}

	for _, memberID := range memberIDs {
		if memberID == msg.UserID {
			continue
		}
		uid := memberID
		mn := &model.Mention{
			ID:              uuid.New(),
			MessageID:       msg.ID,
			ChannelID:       msg.ChannelID,
			MentionedUserID: &uid,
			MentionType:     p.Type,
			CreatedAt:       time.Now(),
		}
		if err := s.repo.Create(ctx, mn); err != nil {
			slog.Error("mention: failed to create channel mention", "error", err)
			continue
		}
		s.notifyMention(mn, msg)
	}
}

func (s *Service) notifyMention(mn *model.Mention, msg *model.Message) {
	if s.broadcast == nil || mn.MentionedUserID == nil {
		return
	}
	// Attach message data so frontend can display mention context
	mn.Message = msg
	payload, _ := json.Marshal(mn)
	event := model.WebSocketEvent{
		Type:      model.EventMentionNew,
		ChannelID: mn.ChannelID.String(),
		Payload:   payload,
	}
	s.broadcast(mn.ChannelID, event)
}

func (s *Service) GetUnreadMentions(ctx context.Context, userID uuid.UUID) ([]model.Mention, error) {
	return s.repo.GetUnreadByUser(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, userID, channelID uuid.UUID) error {
	return s.repo.MarkReadByChannel(ctx, userID, channelID)
}
