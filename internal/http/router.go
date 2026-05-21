package http

import (
	"io/fs"
	"net/http"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/config"
	"github.com/ebachmann/opencode-webchat/internal/http/handlers"
	ghandler "github.com/ebachmann/opencode-webchat/internal/http/handlers/github"
	"github.com/ebachmann/opencode-webchat/internal/opencode"
	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(s *store.Store, jwtManager *auth.JWTManager, staticFS fs.FS, opencodeMgr *opencode.Manager, cfg config.GitHubConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	health := handlers.NewHealthHandler()
	r.Get("/healthz", health.Check)

	authHandler := handlers.NewAuthHandler(jwtManager, s)
	authMiddleware := auth.NewAuthMiddleware(jwtManager, s)
	githubHandler := ghandler.NewHandler(s, cfg)

	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/logout", authHandler.Logout)

	r.With(authMiddleware.Authenticate).Group(func(r chi.Router) {
		r.Get("/api/auth/me", authHandler.Me)

		r.Get("/api/auth/github/login", githubHandler.Login)
		r.Get("/api/auth/github/callback", githubHandler.Callback)
		r.Get("/api/auth/github/status", githubHandler.Status)
		r.Post("/api/auth/github/disconnect", githubHandler.Disconnect)

		projects := handlers.NewProjectsHandler(s)
		r.Get("/api/projects", projects.List)
		r.Post("/api/projects", projects.Create)
		r.Delete("/api/projects/{id}", projects.Delete)
		r.Get("/api/projects/github/repos", projects.ListGitHubRepos)

		r.Route("/api", func(r chi.Router) {
			sessions := handlers.NewSessionsHandler(s)
			r.Get("/sessions", sessions.List)
			r.Post("/sessions", sessions.Create)
			r.Get("/sessions/{id}", sessions.Get)
			r.Patch("/sessions/{id}", sessions.UpdateTitle)
			r.Delete("/sessions/{id}", sessions.Delete)

			messages := handlers.NewMessagesHandler(s)
			r.Get("/sessions/{sessionId}/messages", messages.List)

			settings := handlers.NewSettingsHandler(s, opencodeMgr)
			r.Get("/settings", settings.Get)
			r.Put("/settings", settings.Update)
			r.Get("/settings/models", settings.GetModels)
			r.Get("/settings/providers", settings.GetProviders)

			files := handlers.NewFilesHandler(s)
			r.Get("/files", files.List)
		})
	})

	if staticFS != nil {
		fileServer := http.FileServer(http.FS(staticFS))
		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			filePath := r.URL.Path
			if filePath == "" || filePath == "/" {
				fileServer.ServeHTTP(w, r)
				return
			}

			cleanPath := filePath
			if cleanPath[0] == '/' {
				cleanPath = cleanPath[1:]
			}

			f, err := staticFS.Open(cleanPath)
			if err != nil {
				r.URL.Path = "/"
			} else {
				f.Close()
			}
			fileServer.ServeHTTP(w, r)
		}))
	}

	return r
}