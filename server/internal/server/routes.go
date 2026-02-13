package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/feather-chat/feather/internal/middleware"
)

func (s *Server) setupRoutes() {
	r := s.router

	// WebSocket (no Compress middleware - needs http.Hijacker)
	r.Get("/api/v1/ws", s.wsHandler.ServeHTTP)

	// All other routes use Compress
	r.Group(func(r chi.Router) {
		r.Use(chimiddleware.Compress(5))

		// Health check
		r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})

		// Public routes
		r.Route("/api/v1/auth", func(r chi.Router) {
			r.Post("/register", s.authHandler.Register)
			r.Post("/login", s.authHandler.Login)
			r.Post("/refresh", s.authHandler.RefreshToken)
		})

		// Incoming webhook (token auth in URL)
		r.Post("/api/v1/hooks/{token}", s.webhookHandler.HandleIncoming)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(chimiddleware.Compress(5))
		r.Use(middleware.Auth(s.cfg.JWT.Secret))

		// Auth
		r.Post("/api/v1/auth/logout", s.authHandler.Logout)
		r.Get("/api/v1/auth/me", s.authHandler.Me)

		// Users
		r.Get("/api/v1/users", s.userHandler.List)
		r.Get("/api/v1/users/{userID}", s.userHandler.GetByID)
		r.Patch("/api/v1/users/me", s.userHandler.UpdateProfile)

		// Channels
		r.Route("/api/v1/channels", func(r chi.Router) {
			r.Post("/", s.channelHandler.Create)
			r.Get("/", s.channelHandler.List)

			r.Route("/{channelID}", func(r chi.Router) {
				r.Get("/", s.channelHandler.GetByID)
				r.Patch("/", s.channelHandler.Update)
				r.Delete("/", s.channelHandler.Delete)

				r.Post("/join", s.channelHandler.Join)
				r.Post("/leave", s.channelHandler.Leave)
				r.Post("/members", s.channelHandler.InviteMember)
				r.Get("/members", s.channelHandler.GetMembers)
				r.Post("/read", s.channelHandler.MarkRead)

				// Messages
				r.Post("/messages", s.messageHandler.Create)
				r.Get("/messages", s.messageHandler.List)
				r.Patch("/messages/{messageID}", s.messageHandler.Update)
				r.Delete("/messages/{messageID}", s.messageHandler.Delete)

				// File uploads
				if s.fileHandler != nil {
					r.Post("/files", s.fileHandler.Upload)
				}
			})
		})

		// Threads
		r.Get("/api/v1/messages/{messageID}/thread", s.messageHandler.GetThread)

		// Reactions
		r.Post("/api/v1/messages/{messageID}/reactions", s.reactionHandler.AddReaction)
		r.Delete("/api/v1/messages/{messageID}/reactions/{emoji}", s.reactionHandler.RemoveReaction)

		// Search
		r.Get("/api/v1/search", s.searchHandler.Search)

		// Webhooks management
		r.Route("/api/v1/webhooks", func(r chi.Router) {
			r.Post("/", s.webhookHandler.Create)
			r.Get("/", s.webhookHandler.List)
			r.Delete("/{webhookID}", s.webhookHandler.Delete)
		})

		// File downloads
		if s.fileHandler != nil {
			r.Get("/api/v1/files/{fileID}/download", s.fileHandler.Download)
		}
	})
}
