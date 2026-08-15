package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/kosero/atchannel/backend/internal/config"
	"github.com/kosero/atchannel/backend/internal/handlers"
	"github.com/kosero/atchannel/backend/internal/repositories/postgres"
	"github.com/kosero/atchannel/backend/pkg/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// PostgreSQL.
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return err
	}

	// Redis.
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	defer rdb.Close()

	// Storage.
	st, err := newStorage(cfg)
	if err != nil {
		return err
	}

	// Repositories.
	repos := handlers.Repositories{
		Users:         postgres.NewUserRepo(pool),
		RefreshTokens: postgres.NewRefreshTokenRepo(pool),
		Boards:        postgres.NewBoardRepo(pool),
		Threads:       postgres.NewThreadRepo(pool),
		Replies:       postgres.NewReplyRepo(pool),
		Reports:       postgres.NewReportRepo(pool),
		Bans:          postgres.NewBanRepo(pool),
	}

	router := handlers.NewRouter(cfg, repos, rdb, st)

	server := &http.Server{
		Addr:              cfg.HTTP.Host + ":" + cfg.HTTP.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("api listening", "addr", server.Addr, "env", cfg.App.Env)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func newStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Storage.Provider {
	case "s3":
		return storage.NewS3(storage.S3Config{
			Endpoint:        cfg.Storage.S3.Endpoint,
			AccessKeyID:     cfg.Storage.S3.AccessKeyID,
			SecretAccessKey: cfg.Storage.S3.SecretAccessKey,
			Bucket:          cfg.Storage.S3.Bucket,
			Region:          cfg.Storage.S3.Region,
			UseSSL:          cfg.Storage.S3.UseSSL,
			PublicBaseURL:   cfg.Storage.S3.PublicBaseURL,
		})
	case "local":
		fallthrough
	default:
		return storage.NewLocal(cfg.Storage.Local.BaseDir, cfg.Storage.Local.BaseURL)
	}
}
