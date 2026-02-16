package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/feather-chat/feather/internal/audit"
	"github.com/feather-chat/feather/internal/auth"
	"github.com/feather-chat/feather/internal/call"
	"github.com/feather-chat/feather/internal/channel"
	"github.com/feather-chat/feather/internal/config"
	"github.com/feather-chat/feather/internal/dm"
	"github.com/feather-chat/feather/internal/file"
	"github.com/feather-chat/feather/internal/invitation"
	"github.com/feather-chat/feather/internal/mention"
	"github.com/feather-chat/feather/internal/message"
	"github.com/feather-chat/feather/internal/middleware"
	"github.com/feather-chat/feather/internal/model"
	"github.com/feather-chat/feather/internal/reaction"
	"github.com/feather-chat/feather/internal/search"
	"github.com/feather-chat/feather/internal/user"
	"github.com/feather-chat/feather/internal/usergroup"
	"github.com/feather-chat/feather/internal/webhook"
	"github.com/feather-chat/feather/internal/websocket"
)

type Server struct {
	cfg        *config.Config
	router     *chi.Mux
	httpServer *http.Server
	db         *pgxpool.Pool
	redis      *redis.Client
	hub        *websocket.Hub
	validate   *validator.Validate

	// Handlers
	authHandler       *auth.Handler
	channelHandler    *channel.Handler
	messageHandler    *message.Handler
	reactionHandler   *reaction.Handler
	userHandler       *user.Handler
	webhookHandler    *webhook.Handler
	searchHandler     *search.Handler
	fileHandler       *file.Handler
	wsHandler         *websocket.WSHandler
	invitationHandler *invitation.Handler
	dmHandler         *dm.Handler
	mentionHandler    *mention.Handler
	userGroupHandler  *usergroup.Handler
	callHandler       *call.Handler

	// Services
	channelService *channel.Service
	callService    *call.Service
	auditLogger    *audit.Logger
}

func New(cfg *config.Config, db *pgxpool.Pool, redisClient *redis.Client, fileStorage *file.Storage) *Server {
	s := &Server{
		cfg:      cfg,
		router:   chi.NewRouter(),
		db:       db,
		redis:    redisClient,
		validate: validator.New(),
	}

	s.initServices(fileStorage)
	s.setupMiddleware()
	s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) initServices(fileStorage *file.Storage) {
	// WebSocket hub
	s.hub = websocket.NewHub(s.redis)

	// Audit logger
	s.auditLogger = audit.NewLogger(s.db)

	// Token service
	tokenService := auth.NewTokenService(s.cfg.JWT.Secret, s.cfg.JWT.AccessTTL, s.cfg.JWT.RefreshTTL)

	// Repositories
	authRepo := auth.NewRepository(s.db)
	channelRepo := channel.NewRepository(s.db)
	messageRepo := message.NewRepository(s.db)
	userRepo := user.NewRepository(s.db)
	webhookRepo := webhook.NewRepository(s.db)
	searchRepo := search.NewRepository(s.db)
	invitationRepo := invitation.NewRepository(s.db)
	dmRepo := dm.NewRepository(s.db)
	mentionRepo := mention.NewRepository(s.db)
	userGroupRepo := usergroup.NewRepository(s.db)
	callRepo := call.NewRepository(s.db)

	// Services
	authService := auth.NewService(authRepo, tokenService)
	s.channelService = channel.NewService(channelRepo)
	userService := user.NewService(userRepo)
	searchService := search.NewService(searchRepo)
	invitationService := invitation.NewService(invitationRepo, s.channelService, s.cfg.Server.AppURL)

	// Message service with broadcast
	broadcastFn := func(channelID uuid.UUID, event model.WebSocketEvent) {
		s.hub.BroadcastEvent(channelID, event)
	}
	messageService := message.NewService(messageRepo, s.channelService, broadcastFn)
	reactionService := reaction.NewService(s.db, broadcastFn)

	// Mention service (processes @mentions in messages)
	mentionService := mention.NewService(mentionRepo, broadcastFn)
	messageService.SetMentionProcessor(mentionService)

	// DM service with subscribe callback
	subscribeFn := func(userID uuid.UUID, channelID uuid.UUID) {
		s.hub.SubscribeUserToChannel(userID, channelID)
	}
	dmService := dm.NewService(dmRepo, broadcastFn, subscribeFn)

	// User group service
	userGroupService := usergroup.NewService(userGroupRepo)

	// Call service
	sendToUserFn := func(userID uuid.UUID, data []byte) {
		s.hub.SendToUser(userID, data)
	}
	s.callService = call.NewService(callRepo, broadcastFn, sendToUserFn)
	s.callService.SetMemberChecker(s.channelService)

	// Set up call event handler on the hub
	s.hub.SetCallHandler(func(userID uuid.UUID, event model.WebSocketEvent) {
		s.handleCallWSEvent(userID, event)
	})

	// Webhook service (bot user ID will be set after seeding)
	webhookService := webhook.NewService(webhookRepo, uuid.Nil, nil)

	// Handlers
	s.authHandler = auth.NewHandler(authService, s.validate, s.channelService, s.cfg.OAuth.GoogleClientID)
	s.authHandler.SetInvitationAcceptor(invitationService)
	s.channelHandler = channel.NewHandler(s.channelService, s.validate)
	s.messageHandler = message.NewHandler(messageService, s.validate)
	s.reactionHandler = reaction.NewHandler(reactionService, s.validate)
	s.userHandler = user.NewHandler(userService, s.validate)
	s.webhookHandler = webhook.NewHandler(webhookService, s.validate)
	s.searchHandler = search.NewHandler(searchService)
	s.wsHandler = websocket.NewHandler(s.hub, s.cfg.JWT.Secret, userService)
	s.invitationHandler = invitation.NewHandler(invitationService, s.validate)
	s.dmHandler = dm.NewHandler(dmService, s.validate)
	s.mentionHandler = mention.NewHandler(mentionService, s.validate)
	s.userGroupHandler = usergroup.NewHandler(userGroupService, s.validate)
	s.callHandler = call.NewHandler(s.callService, s.cfg.WebRTC)

	if fileStorage != nil {
		s.fileHandler = file.NewHandler(fileStorage, s.db, s.cfg.Upload.MaxSize)
	}
}

func (s *Server) handleCallWSEvent(userID uuid.UUID, event model.WebSocketEvent) {
	ctx := context.Background()

	switch event.Type {
	case model.EventCallInitiate:
		var req model.InitiateCallRequest
		if err := json.Unmarshal(event.Payload, &req); err != nil {
			return
		}
		s.callService.Initiate(ctx, req, userID)

	case model.EventCallAccept:
		var payload struct {
			CallID uuid.UUID `json:"call_id"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return
		}
		s.callService.Accept(ctx, payload.CallID, userID)

	case model.EventCallDecline:
		var payload struct {
			CallID uuid.UUID `json:"call_id"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return
		}
		s.callService.Decline(ctx, payload.CallID, userID)

	case model.EventCallHangup:
		var payload struct {
			CallID uuid.UUID `json:"call_id"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return
		}
		s.callService.Hangup(ctx, payload.CallID, userID)

	case model.EventCallOffer, model.EventCallAnswer, model.EventCallICECandidate:
		var msg call.SignalingMessage
		if err := json.Unmarshal(event.Payload, &msg); err != nil {
			return
		}
		msg.FromUser = userID
		s.callService.RelaySignaling(event.Type, msg)
	}
}

func (s *Server) setupMiddleware() {
	s.router.Use(chimiddleware.RequestID)
	s.router.Use(chimiddleware.RealIP)
	s.router.Use(middleware.Logging)
	s.router.Use(chimiddleware.Recoverer)
	s.router.Use(cors.Handler(middleware.CORS()))
	// Compress is applied per-route group to avoid breaking WebSocket hijack
}

func (s *Server) Start() error {
	// Start WebSocket hub
	go s.hub.Run()

	// Seed default channel
	ctx := context.Background()
	if _, err := s.channelService.SeedDefaultChannel(ctx); err != nil {
		slog.Warn("failed to seed default channel", "error", err)
	}

	// Recover any calls stuck in ringing state from a prior shutdown
	s.callService.RecoverStaleCalls(ctx)

	slog.Info("server starting", "port", s.cfg.Server.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("server shutting down")
	s.hub.Stop()
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Router() *chi.Mux {
	return s.router
}
