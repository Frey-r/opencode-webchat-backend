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
	"strings"
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
	router := httpkit.NewRouter(s, jwtManager, staticFS, ocManager, cfg.GitHub)

	wsHandler := &wsHandler{
		hub:        hub,
		jwtManager: jwtManager,
		store:      s,
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
	store      *store.Store
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[WS] incoming request: %s %s from %s", r.Method, r.URL.String(), r.RemoteAddr)

	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		sessionIDStr = r.URL.Query().Get("session")
	}
	if sessionIDStr == "" {
		log.Printf("[WS] no session_id provided")
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	sessionID := 0
	var err error
	if sessionIDStr == "default" {
		sessions, dbErr := h.store.GetSessionsByProject(r.Context(), 1)
		if dbErr == nil && len(sessions) > 0 {
			sessionID = sessions[0].ID
		} else {
			sess, createErr := h.store.CreateSession(r.Context(), 1, "Default Session")
			if createErr == nil && sess != nil {
				sessionID = sess.ID
			} else {
				sessionID = 1
			}
		}
	} else {
		sessionID, err = strconv.Atoi(sessionIDStr)
		if err != nil {
			sessionID = 1
		}
	}

	tokenStr := ""
	cookieName := h.jwtManager.CookieName()
	if cookie, err := r.Cookie(cookieName); err == nil && cookie.Value != "" {
		tokenStr = cookie.Value
		log.Printf("[WS] auth via cookie")
	} else if t := r.URL.Query().Get("token"); t != "" {
		tokenStr = t
		log.Printf("[WS] auth via query token")
	} else if authHeader := r.Header.Get("Authorization"); authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		log.Printf("[WS] auth via bearer header")
	} else {
		log.Printf("[WS] no auth found: cookie_err=%v, token=%s, auth=%s", r.Cookies(), r.URL.Query().Get("token"), r.Header.Get("Authorization"))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtManager.Validate(tokenStr)
	if err != nil {
		log.Printf("[WS] invalid token: %v", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	log.Printf("[WS] authenticated user %d, session %d", claims.UserID, sessionID)

	workDir := ""
	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr != "" {
		if pid, err := strconv.Atoi(projectIDStr); err == nil {
			if p, err := h.store.GetProject(r.Context(), pid); err == nil && p != nil {
				workDir = p.RootPath
				log.Printf("[WS] project %d workDir=%s", pid, workDir)
			}
		}
	}

	h.hub.HandleWS(w, r, sessionID, claims.UserID, workDir)
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