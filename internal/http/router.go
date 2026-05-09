package http

import (
	"io/fs"
	"net/http"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/http/handlers"
	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(s *store.Store, jwtManager *auth.JWTManager, staticFS fs.FS) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	health := handlers.NewHealthHandler()
	r.Get("/healthz", health.Check)

	authHandler := handlers.NewAuthHandler(jwtManager, s)
	authMiddleware := auth.NewAuthMiddleware(jwtManager, s)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/logout", authHandler.Logout)

		r.With(authMiddleware.Authenticate).Group(func(r chi.Router) {
			r.Get("/auth/me", authHandler.Me)

			projects := handlers.NewProjectsHandler(s)
			r.Get("/projects", projects.List)

			sessions := handlers.NewSessionsHandler(s)
			r.Get("/sessions", sessions.List)
			r.Post("/sessions", sessions.Create)
			r.Get("/sessions/{id}", sessions.Get)
			r.Patch("/sessions/{id}", sessions.UpdateTitle)
			r.Delete("/sessions/{id}", sessions.Delete)

			messages := handlers.NewMessagesHandler(s)
			r.Get("/sessions/{sessionId}/messages", messages.List)
		})
	})

	if staticFS != nil {
		subFS, err := fs.Sub(staticFS, "web/dist")
		if err == nil {
			r.Handle("/*", http.FileServer(http.FS(subFS)))
		}
	}

	return r
}