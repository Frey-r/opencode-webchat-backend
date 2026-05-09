package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/config"
	httpkit "github.com/ebachmann/opencode-webchat/internal/http"
	"github.com/ebachmann/opencode-webchat/internal/opencode"
	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/ebachmann/opencode-webchat/internal/ws"
)

//go:embed web/dist
var embeddedFS embed.FS

func main() {
	migrate := flag.Bool("migrate", false, "run migrations on startup")
	flag.Parse()

	cfg := config.Load()

	s, err := store.New(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer s.Close()

	if *migrate {
		log.Println("note: migrations should be run with goose CLI")
	}

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.CookieName)

	ocManager := opencode.NewManager(cfg.OpenCode.BinaryPath)
	defer ocManager.StopAll()

	hub := ws.NewHub(ocManager, s)
	hubCtx, hubCancel := context.WithCancel(context.Background())
	go hub.Run(hubCtx)

	staticFS, _ := fs.Sub(embeddedFS, "web/dist")
	router := httpkit.NewRouter(s, jwtManager, staticFS)

	wsHandler := &wsHandler{
		hub:        hub,
		jwtManager: jwtManager,
	}
	router.Handle("/ws", wsHandler)

	srv := &http.Server{
		Addr:         cfg.Address(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		hubCancel()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	createDemoUser(s)

	log.Printf("server starting on %s", cfg.Address())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

type wsHandler struct {
	hub        *ws.Hub
	jwtManager *auth.JWTManager
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		http.Error(w, "invalid session_id", http.StatusBadRequest)
		return
	}

	tokenStr := h.jwtManager.CookieName()
	if cookie, err := r.Cookie(tokenStr); err == nil && cookie.Value != "" {
		tokenStr = cookie.Value
	} else {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtManager.Validate(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	h.hub.HandleWS(w, r, sessionID, claims.UserID)
}

func createDemoUser(s *store.Store) {
	user, err := s.GetUserByUsername(context.Background(), "admin")
	if err != nil || user != nil {
		return
	}

	hash, _ := auth.HashPassword("admin123")
	s.CreateUser(context.Background(), "admin", hash)
	log.Println("demo user created: admin / admin123")
}