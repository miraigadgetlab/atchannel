package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"github.com/kosero/atchannel/backend/internal/config"
	"github.com/kosero/atchannel/backend/internal/middleware"
	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
	"github.com/kosero/atchannel/backend/internal/services"
	"github.com/kosero/atchannel/backend/pkg/storage"
)

// Repositories bundles the repository interfaces the router needs.
type Repositories struct {
	Users         repositories.UserRepository
	RefreshTokens repositories.RefreshTokenRepository
	Boards        repositories.BoardRepository
	Threads       repositories.ThreadRepository
	Replies       repositories.ReplyRepository
	Reports       repositories.ReportRepository
	Bans          repositories.BanRepository
}

// NewRouter wires repositories, services and handlers into an http.Handler.
func NewRouter(cfg *config.Config, repos Repositories, rdb *redis.Client, st storage.Storage) *chi.Mux {
	// Services.
	authSvc := services.NewAuthService(repos.Users, repos.RefreshTokens, &cfg.Auth)
	userSvc := services.NewUserService(repos.Users)
	boardSvc := services.NewBoardService(repos.Boards)
	uploadSvc := services.NewUploadService(st, cfg.Upload.MaxSizeBytes, cfg.Upload.AllowedTypes, cfg.Storage.Local.BaseURL)
	threadSvc := services.NewThreadService(repos.Threads, repos.Replies, repos.Reports, repos.Boards, repos.Bans, uploadSvc)
	modSvc := services.NewModerationService(repos.Reports, repos.Bans, repos.Threads, repos.Replies, repos.Users)

	// Handlers.
	authHandler := NewAuthHandler(authSvc, &cfg.Auth)
	boardHandler := NewBoardHandler(boardSvc)
	threadHandler := NewThreadHandler(threadSvc, userSvc)
	userHandler := NewUserHandler(userSvc)
	uploadHandler := NewUploadHandler(uploadSvc)
	reportHandler := NewReportHandler(modSvc, userSvc)
	adminHandler := NewAdminHandler(modSvc, userSvc)

	// Rate limiters.
	postLimiter := middleware.NewSlidingWindowLimiter(rdb, time.Minute, 10)
	refreshLimiter := middleware.NewSlidingWindowLimiter(rdb, time.Minute, 30)

	auth := middleware.NewAuth(authSvc)

	router := chi.NewRouter()
	router.Use(chimw.RealIP)
	router.Use(middleware.RequestLogging)
	router.Use(middleware.Recoverer)

	// NOTE: chi.RedirectSlashes has a known open-redirect issue
	// (GO-2026-4316); we deliberately do not mount it.

	// Serve locally stored files only for the local provider; when S3 is
	// used, objects are served by the bucket/CDN and no route is mounted.
	if cfg.Storage.Provider == "local" {
		fileHandler := NewFileHandler(st)
		router.Get("/files/*", fileHandler.Serve)
	}

	router.Route("/api", func(api chi.Router) {
		api.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		api.Get("/boards", boardHandler.List)

		api.Group(func(gr chi.Router) {
			gr.Use(auth.OptionalAuth)
			gr.Get("/boards/{board}/threads", threadHandler.ListByBoard)
			gr.Get("/threads/{id}", threadHandler.Get)
			gr.Get("/users/{username}", userHandler.GetByUsername)
		})

		api.Group(func(gr chi.Router) {
			gr.Use(auth.RequireAuth)
			gr.With(postLimiter.Middleware(ipOrUserKey("thread"))).Post("/boards/{board}/threads", threadHandler.Create)
			gr.With(postLimiter.Middleware(ipOrUserKey("reply"))).Post("/threads/{id}/replies", threadHandler.Reply)
			gr.With(postLimiter.Middleware(ipOrUserKey("upload"))).Post("/upload", uploadHandler.Upload)
			gr.Patch("/users/me", userHandler.UpdateMe)
			gr.Post("/reports", reportHandler.Create)
		})

		api.Route("/auth", func(authR chi.Router) {
			authR.Post("/register", authHandler.Register)
			authR.Post("/login", authHandler.Login)
			authR.With(refreshLimiter.Middleware(ipKey)).Post("/refresh", authHandler.Refresh)
			authR.Post("/logout", authHandler.Logout)
		})

		api.Group(func(admin chi.Router) {
			admin.Use(auth.RequireAuth)
			admin.Use(middleware.RequireRole(models.RoleMod))
			admin.Post("/admin/threads/{id}/delete", adminHandler.DeleteThread)
			admin.Post("/admin/replies/{id}/delete", adminHandler.DeleteReply)
			admin.Get("/admin/reports", adminHandler.ListReports)
			admin.Post("/admin/reports/{id}/resolve", adminHandler.ResolveReport)
			admin.Post("/admin/bans", adminHandler.CreateBan)
			admin.Get("/admin/bans", adminHandler.ListBans)
		})
	})

	return router
}

func ipKey(r *http.Request) string {
	return "ip:" + r.RemoteAddr
}

func ipOrUserKey(kind string) func(r *http.Request) string {
	return func(r *http.Request) string {
		user := middleware.ContextUserID(r.Context())
		if user != "" {
			return kind + ":user:" + user
		}
		return kind + ":ip:" + r.RemoteAddr
	}
}
