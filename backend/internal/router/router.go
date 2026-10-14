package router

import (
	"time"

	"atchannel-backend/internal/config"
	"atchannel-backend/internal/handler"
	"atchannel-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/golang-jwt/jwt/v5"
)

type RouterDeps struct {
	Config              *config.Config
	AuthHandler         *handler.AuthHandler
	ChannelHandler      *handler.ChannelHandler
	PostHandler         *handler.PostHandler
	CommentHandler      *handler.CommentHandler
	AdminHandler        *handler.AdminHandler
	ProfileHandler      *handler.ProfileHandler
	AccountHandler      *handler.AccountHandler
	NotificationHandler *handler.NotificationHandler
}

func SetupRoutes(app *fiber.App, deps RouterDeps) {
	// Panic recovery is the outermost layer, so a bug in any middleware or
	// handler becomes a logged 500 (via the app ErrorHandler) instead of a
	// crashed process.
	app.Use(recover.New())

	// Stamp every request (and every future log line) with a correlation id
	// before anything that might fail runs. See middleware.RequestID.
	app.Use(middleware.RequestID())

	// Cap request lifetime so a slow query cannot pin a DB connection; the
	// deadline rides along on c.Context(), which the services pass to GORM.
	app.Use(middleware.Timeout(15 * time.Second))

	// CORS first so preflight OPTIONS is answered before any auth guard.
	app.Use(cors.New(cors.Config{
		AllowOrigins: deps.Config.AllowedOrigins,
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodHead,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
		},
		AllowHeaders: []string{
			fiber.HeaderOrigin,
			fiber.HeaderContentType,
			fiber.HeaderAccept,
			fiber.HeaderAuthorization,
		},
		MaxAge: 3600,
	}))

	app.Use(middleware.RequestLogger())

	// Uploaded files (avatars) are served statically from local disk.
	if deps.Config.UploadDir != "" {
		app.Use("/uploads", static.New(deps.Config.UploadDir))
	}

	api := app.Group("/api/v1")

	authConfig := middleware.AuthConfig{
		SecretKey:     deps.Config.JWTSecret,
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        deps.Config.JWTIssuer,
	}
	authGuard := middleware.NewAuthMiddleware(authConfig)
	// Public reads use the same parsing but tolerate missing/bad tokens, so a
	// logged-in user still gets liked_by_me without breaking anonymous access.
	optionalAuth := middleware.OptionalAuth(authConfig)

	// Brute force protection on the public auth endpoints. Register counts
	// every attempt (mass signups are the attack), login only counts
	// failures (a legit user must never lock themselves out by succeeding).
	registerLimit := middleware.NewAuthRateLimiter(deps.Config.RateLimitMax, deps.Config.RateLimitWindow, false)
	loginLimit := middleware.NewAuthRateLimiter(deps.Config.RateLimitMax, deps.Config.RateLimitWindow, true)
	// Password-reset requests are counted even when they succeed: the point is
	// to cap how fast an address can be probed / spammed.
	forgotLimit := middleware.NewAuthRateLimiter(deps.Config.RateLimitMax, deps.Config.RateLimitWindow, false)

	// Content creation is throttled per user, not per IP: it caps how fast
	// a single account can flood the site regardless of where it connects
	// from.
	writeLimit := middleware.NewWriteRateLimiter(deps.Config.WriteLimitMax, deps.Config.WriteLimitWindow)

	// Health is unauthenticated and dependency-free: it reports process
	// liveness only, which is exactly what a container healthcheck needs.
	// (Readiness, i.e. "is the database up", is the compose healthcheck's
	// job via the postgres service itself.)
	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Public
	api.Post("/auth/register", registerLimit, deps.AuthHandler.Register)
	api.Post("/auth/login", loginLimit, deps.AuthHandler.Login)
	api.Post("/auth/refresh", deps.AuthHandler.Refresh)
	api.Post("/auth/logout", deps.AuthHandler.Logout)
	api.Post("/auth/verify-email", deps.AccountHandler.VerifyEmail)
	api.Post("/auth/forgot-password", forgotLimit, deps.AccountHandler.ForgotPassword)
	api.Post("/auth/reset-password", forgotLimit, deps.AccountHandler.ResetPassword)
	api.Post("/auth/reset-password-code", forgotLimit, deps.AccountHandler.ResetPasswordWithCode)

	api.Get("/channels", deps.ChannelHandler.List)
	api.Get("/channels/:id", deps.ChannelHandler.GetByID)
	api.Get("/posts", optionalAuth, deps.PostHandler.List)
	api.Get("/posts/:id", optionalAuth, deps.PostHandler.GetByID)
	api.Get("/posts/:id/comments", deps.CommentHandler.List)
	api.Get("/users", deps.ProfileHandler.ListChannelers)
	api.Get("/users/:id", deps.ProfileHandler.GetByID)

	// Authenticated
	protected := api.Group("", authGuard)

	protected.Get("/me", deps.ProfileHandler.Me)
	protected.Put("/me", deps.ProfileHandler.Update)
	protected.Put("/me/password", deps.ProfileHandler.ChangePassword)
	protected.Post("/me/verify-email", deps.AccountHandler.ResendVerification)
	protected.Get("/me/recovery-codes", deps.AccountHandler.RecoveryCodeStatus)
	protected.Post("/me/recovery-codes", deps.AccountHandler.RecoveryCodes)
	protected.Post("/me/avatar", writeLimit, deps.ProfileHandler.UploadAvatar)
	protected.Post("/auth/logout-all", deps.AuthHandler.LogoutAll)
	protected.Post("/channels", writeLimit, deps.ChannelHandler.Create)
	protected.Put("/channels/:id", writeLimit, deps.ChannelHandler.Update)
	// Creator or admin — the handler enforces ownership (see ChannelHandler.Delete).
	protected.Delete("/channels/:id", writeLimit, deps.ChannelHandler.Delete)
	protected.Post("/posts", writeLimit, deps.PostHandler.Create)
	protected.Put("/posts/:id", writeLimit, deps.PostHandler.Update)
	protected.Delete("/posts/:id", writeLimit, deps.PostHandler.Delete)
	protected.Post("/posts/:id/comments", writeLimit, deps.CommentHandler.Create)
	protected.Delete("/comments/:commentID", writeLimit, deps.CommentHandler.Delete)
	protected.Post("/posts/:id/like", writeLimit, deps.PostHandler.Like)
	protected.Delete("/posts/:id/like", writeLimit, deps.PostHandler.Unlike)

	protected.Get("/notifications", deps.NotificationHandler.List)
	protected.Get("/notifications/unread-count", deps.NotificationHandler.UnreadCount)
	protected.Post("/notifications/read", deps.NotificationHandler.MarkAllRead)

	// Admin only
	admin := api.Group("", authGuard, middleware.RequireRole("admin"))

	admin.Get("/admin/users", deps.AdminHandler.ListUsers)
	admin.Get("/admin/stats", deps.AdminHandler.Stats)
	admin.Put("/admin/users/:id/roles", deps.AdminHandler.SetRoles)
	admin.Put("/admin/users/:id/ban", deps.AdminHandler.SetBanned)
	admin.Delete("/admin/users/:id", deps.AdminHandler.DeleteUser)
	admin.Post("/admin/users/:id/recovery-codes", deps.AdminHandler.GenerateRecoveryCodes)

	// Moderation: admins and moderators may pin/lock threads.
	mod := api.Group("", authGuard, middleware.RequireRole("admin", "moderator"))

	mod.Post("/posts/:id/pin", deps.PostHandler.Pin)
	mod.Delete("/posts/:id/pin", deps.PostHandler.Unpin)
	mod.Post("/posts/:id/lock", deps.PostHandler.Lock)
	mod.Delete("/posts/:id/lock", deps.PostHandler.Unlock)
}
