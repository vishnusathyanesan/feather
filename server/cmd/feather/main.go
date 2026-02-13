package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/feather-chat/feather/internal/config"
	"github.com/feather-chat/feather/internal/database"
	"github.com/feather-chat/feather/internal/file"
	"github.com/feather-chat/feather/internal/server"
)

func main() {
	migrateUp := flag.Bool("migrate-up", false, "Run migrations up")
	migrateDown := flag.Bool("migrate-down", false, "Run migrations down")
	flag.Parse()

	// Setup structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Run migrations
	migrationsPath := "migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = "/app/migrations"
	}

	if *migrateUp {
		if err := database.RunMigrations(cfg.Database.URL, migrationsPath); err != nil {
			slog.Error("migration up failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migrations applied successfully")
		return
	}

	if *migrateDown {
		if err := database.RollbackMigrations(cfg.Database.URL, migrationsPath); err != nil {
			slog.Error("migration down failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migrations rolled back successfully")
		return
	}

	// Auto-run migrations on startup
	if err := database.RunMigrations(cfg.Database.URL, migrationsPath); err != nil {
		slog.Error("auto-migration failed", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Init PostgreSQL
	db, err := database.NewPostgresPool(ctx, cfg.Database.URL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to PostgreSQL")

	// Init Redis
	redisClient, err := database.NewRedisClient(ctx, cfg.Redis.URL)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("connected to Redis")

	// Init MinIO file storage
	var fileStorage *file.Storage
	if cfg.MinIO.Endpoint != "" {
		fileStorage, err = file.NewStorage(
			cfg.MinIO.Endpoint, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey,
			cfg.MinIO.UseSSL, cfg.MinIO.Bucket,
		)
		if err != nil {
			slog.Warn("failed to init file storage", "error", err)
		} else {
			if err := fileStorage.EnsureBucket(ctx); err != nil {
				slog.Warn("failed to ensure bucket", "error", err)
			} else {
				slog.Info("connected to MinIO", "bucket", cfg.MinIO.Bucket)
			}
		}
	}

	// Create and start server
	srv := server.New(cfg, db, redisClient, fileStorage)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
