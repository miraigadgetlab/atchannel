package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"atchannel-backend/internal/config"
	"atchannel-backend/internal/db"
	"atchannel-backend/internal/handler"
	"atchannel-backend/internal/middleware"
	"atchannel-backend/internal/router"
	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg := config.LoadConfig()

	db.Connect()
	db.Migrate()

	tokenService := service.NewTokenService(cfg.JWTSecret, "atchannel")
	userService := service.NewUserService(db.DB)
	channelService := service.NewChannelService(db.DB)
	postService := service.NewPostService(db.DB)
	commentService := service.NewCommentService(db.DB)
	statsService := service.NewStatsService(db.DB)
	sessionService := service.NewSessionService(db.DB)

	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		admin, err := userService.EnsureAdmin(context.Background(), cfg.AdminName, cfg.AdminEmail, cfg.AdminPassword)
		if err != nil {
			log.Fatalf("admin bootstrap failed: %v", err)
		}
		log.Printf("admin account ready: %s <%s> roles=%v", admin.Name, admin.Email, admin.Roles)
	}

	accountLimiter := middleware.NewAccountLimiter(cfg.AccountLockMax, cfg.AccountLockWindow)
	authHandler := handler.NewAuthHandler(tokenService, userService, sessionService, accountLimiter)
	channelHandler := handler.NewChannelHandler(channelService, postService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	adminHandler := handler.NewAdminHandler(userService, statsService, sessionService)
	profileHandler := handler.NewProfileHandler(userService, sessionService)

	fiberCfg := fiber.Config{
		// A single abusive client must not be able to pin memory with a
		// giant body. 1MB comfortably fits the longest allowed post.
		BodyLimit: 1 << 20,
	}

	// Behind a reverse proxy every request arrives from the proxy's address,
	// so per-IP rate limits would collapse into a single bucket — or worse,
	// be escapable by forging X-Forwarded-For. Trust is opt-in: only the
	// listed proxies get their forwarded header believed.
	if len(cfg.TrustedProxies) > 0 {
		fiberCfg.TrustProxy = true
		fiberCfg.ProxyHeader = fiber.HeaderXForwardedFor
		fiberCfg.TrustProxyConfig.Proxies = cfg.TrustedProxies
	}

	app := fiber.New(fiberCfg)

	router.SetupRoutes(app, router.RouterDeps{
		Config:         cfg,
		AuthHandler:    authHandler,
		ChannelHandler: channelHandler,
		PostHandler:    postHandler,
		CommentHandler: commentHandler,
		AdminHandler:   adminHandler,
		ProfileHandler: profileHandler,
	})

	// Graceful shutdown: SIGTERM/SIGINT lets in-flight requests finish
	// instead of dropping them when the container stops.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		log.Print("shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("listening on %s", cfg.Port)
	if err := app.Listen(cfg.Port); err != nil && !errors.Is(err, net.ErrClosed) {
		log.Fatalf("server error: %v", err)
	}

	log.Print("stopped cleanly")
}
