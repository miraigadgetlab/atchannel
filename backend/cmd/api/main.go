package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"atchannel-backend/internal/config"
	"atchannel-backend/internal/db"
	"atchannel-backend/internal/handler"
	"atchannel-backend/internal/mailer"
	"atchannel-backend/internal/middleware"
	"atchannel-backend/internal/router"
	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg := config.LoadConfig()

	db.Connect()
	defer db.Close()
	db.Migrate()

	// Avatar uploads go to local disk; make sure it exists before the first
	// upload (and before the static middleware starts serving it).
	if cfg.UploadDir != "" {
		if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
			log.Fatalf("could not create upload dir %q: %v", cfg.UploadDir, err)
		}
	}

	tokenService := service.NewTokenService(cfg.JWTSecret, cfg.JWTIssuer)
	userService := service.NewUserService(db.DB)
	channelService := service.NewChannelService(db.DB)
	postService := service.NewPostService(db.DB)
	notificationService := service.NewNotificationService(db.DB)
	commentService := service.NewCommentService(db.DB, notificationService)
	statsService := service.NewStatsService(db.DB)
	sessionService := service.NewSessionService(db.DB)

	// Transactional email: SMTP when configured, otherwise log-only so the
	// password reset flow works in development.
	accountMailer := mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.MailFrom)
	accountService := service.NewAccountService(db.DB, userService, sessionService, accountMailer, cfg.AppBaseURL, cfg.ResetTokenTTL)

	// Dead refresh tokens are pruned on a schedule: Record only clears the
	// rows of users who log in again, so abandoned accounts would otherwise
	// accumulate expired sessions forever.
	go runTokenPruner(sessionService)

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
	postHandler := handler.NewPostHandler(postService, notificationService)
	commentHandler := handler.NewCommentHandler(commentService)
	adminHandler := handler.NewAdminHandler(userService, statsService, sessionService, accountService)
	profileHandler := handler.NewProfileHandler(userService, sessionService, cfg.UploadDir, cfg.MaxAvatarBytes)
	accountHandler := handler.NewAccountHandler(accountService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	fiberCfg := fiber.Config{
		// A single abusive client must not be able to pin memory with a
		// giant body. 3MB leaves room for the multipart envelope around an
		// at-most-1MB avatar while still capping abusive uploads.
		BodyLimit: 3 << 20,
		// Keep every error in the same {"error","message"} envelope the
		// handlers use. Fiber errors (405, body limit, ...) keep their
		// status; anything else is a logged 500 with no internal detail.
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Internal Server Error"

			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
				message = fe.Message
			} else {
				if id, ok := c.Locals(middleware.RequestIDKey).(string); ok && id != "" {
					log.Printf("unhandled request error (%s %s) id=%s: %v", c.Method(), c.Path(), id, err)
				} else {
					log.Printf("unhandled request error (%s %s): %v", c.Method(), c.Path(), err)
				}
			}

			return c.Status(code).JSON(fiber.Map{
				"error":   http.StatusText(code),
				"message": message,
			})
		},
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
		Config:              cfg,
		AuthHandler:         authHandler,
		ChannelHandler:      channelHandler,
		PostHandler:         postHandler,
		CommentHandler:      commentHandler,
		AdminHandler:        adminHandler,
		ProfileHandler:      profileHandler,
		AccountHandler:      accountHandler,
		NotificationHandler: notificationHandler,
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

// runTokenPruner deletes expired refresh tokens every few hours until the
// process exits.
func runTokenPruner(sessions *service.SessionService) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		pruned, err := sessions.PruneExpired(ctx)
		cancel()

		if err != nil {
			log.Printf("refresh token prune failed: %v", err)
			continue
		}
		if pruned > 0 {
			log.Printf("pruned %d expired refresh tokens", pruned)
		}
	}
}
