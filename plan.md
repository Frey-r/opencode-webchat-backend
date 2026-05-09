# OpenCode WebChat Backend — Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Backend Go service that bridges a Vue 3 web frontend with the `opencode` CLI via PTY, with WebSocket real-time communication and PostgreSQL persistence.

**Architecture:** Single Go binary with embedded Vue SPA (`embed.FS`). REST API for CRUD operations, WebSocket hub for bidirectional streaming with opencode PTY processes. JWT auth with httpOnly cookies. PostgreSQL for sessions/messages/users.

**Tech Stack:** Go 1.22+, `chi` router, `gorilla/websocket`, `creack/pty`, `golang-jwt/jwt/v5`, `pgx/v5`, `goose`, `lumberjack`.

---

## Phase 1: Project Scaffold & Config

### Task 1: Create project directory structure

**Files:**
- Create: `cmd/server/main.go`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `.env.example`
- Create: `go.mod`
- Create: `Makefile`

**Step 1: Create `go.mod`**

```go
module github.com/ebachmann/opencode-webchat

go 1.22
```

**Step 2: Create `internal/config/config.go`**

```go
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OpenCode OpenCodeConfig
}

type ServerConfig struct {
	Host string
	Port string
	Dir  string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	Expiry     time.Duration
	CookieName string
}

type OpenCodeConfig struct {
	BinaryPath string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("HOST", "0.0.0.0"),
			Port: getEnv("PORT", "8080"),
			Dir:  getEnv("SERVE_DIR", "./web/dist"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/opencode_webchat?sslmode=disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "change-me-in-production"),
			Expiry:     duration(getEnv("JWT_EXPIRY", "24h")),
			CookieName: getEnv("JWT_COOKIE_NAME", "owc_token"),
		},
		OpenCode: OpenCodeConfig{
			BinaryPath: getEnv("OPENCODE_BINARY", "opencode"),
		},
	}
}

func getEnv(key, default_ string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return default_
}

func duration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func (c *Config) Address() string {
	return c.Server.Host + ":" + c.Server.Port
}
```

**Step 3: Create `internal/config/config_test.go`**

```go
package config

import (
	"os"
	"testing"
)

func TestConfigLoadDefaults(t *testing.T) {
	os.Clearenv()
	cfg := Load()

	if cfg.Server.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Server.Port)
	}
	if cfg.JWT.Expiry != 24*time.Hour {
		t.Errorf("expected default expiry 24h, got %v", cfg.JWT.Expiry)
	}
}

func TestConfigAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Host: "127.0.0.1", Port: "3000"},
	}
	if cfg.Address() != "127.0.0.1:3000" {
		t.Errorf("expected 127.0.0.1:3000, got %s", cfg.Address())
	}
}
```

**Step 4: Create `.env.example`**

```env
HOST=0.0.0.0
PORT=8080
SERVE_DIR=./web/dist
DATABASE_URL=postgres://postgres:postgres@localhost:5432/opencode_webchat?sslmode=disable
JWT_SECRET=change-me-in-production
JWT_EXPIRY=24h
JWT_COOKIE_NAME=owc_token
OPENCODE_BINARY=opencode
```

**Step 5: Create `Makefile`**

```makefile
.PHONY: build run test migrate dev clean

BINARY=opencode-webchat
MAIN=cmd/server/main.go

build:
	go build -o $(BINARY) $(MAIN)

run: build
	./$(BINARY)

test:
	go test ./...

dev:
	go run $(MAIN)

clean:
	rm -f $(BINARY)
```

---

## Phase 2: Database & Migrations

### Task 2: Set up goose migrations

**Files:**
- Create: `migrations/000001_init_schema.up.sql`
- Create: `migrations/000001_init_schema.down.sql`
- Create: `internal/store/store.go`
- Create: `internal/store/user.go`
- Create: `internal/store/session.go`
- Create: `internal/store/message.go`
- Create: `internal/store/store_test.go`

**Step 1: Create `migrations/000001_init_schema.up.sql`**

```sql
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    root_path VARCHAR(1024) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL DEFAULT 'New Session',
    opencode_session_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    session_id INTEGER REFERENCES sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL, -- 'user' or 'assistant'
    content TEXT NOT NULL DEFAULT '',
    tool_calls JSONB DEFAULT '[]',
    status VARCHAR(20) DEFAULT 'completed', -- 'completed', 'interrupted', 'error'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Attachments table
CREATE TABLE IF NOT EXISTS attachments (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES messages(id) ON DELETE CASCADE,
    file_path VARCHAR(1024) NOT NULL,
    file_hash VARCHAR(64)
);

-- Audit log table
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    action VARCHAR(100) NOT NULL,
    payload JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_fts ON messages USING gin(to_tsvector('simple', content));
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action);

-- Full-text search extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_messages_trgm ON messages USING gin(content gin_trgm_ops);
```

**Step 2: Create `migrations/000001_init_schema.down.sql`**

```sql
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
```

**Step 3: Create `internal/store/store.go`**

```go
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}
```

**Step 4: Create `internal/store/user.go`**

```go
package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username`,
		username, passwordHash,
	).Scan(&user.ID, &user.Username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```

**Step 5: Create `internal/store/session.go`**

```go
package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Session struct {
	ID               int
	ProjectID        int
	Title            string
	OpenCodeSessionID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	IsActive         bool
}

func (s *Store) CreateSession(ctx context.Context, projectID int, title string) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx,
		`INSERT INTO sessions (project_id, title) VALUES ($1, $2)
		 RETURNING id, project_id, title, opencode_session_id, created_at, updated_at, is_active`,
		projectID, title,
	).Scan(&session.ID, &session.ProjectID, &session.Title, &session.OpenCodeSessionID, &session.CreatedAt, &session.UpdatedAt, &session.IsActive)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *Store) GetSessionsByProject(ctx context.Context, projectID int) ([]Session, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, title, opencode_session_id, created_at, updated_at, is_active
		 FROM sessions WHERE project_id = $1 AND is_active = TRUE ORDER BY updated_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var session Session
		if err := rows.Scan(&session.ID, &session.ProjectID, &session.Title, &session.OpenCodeSessionID, &session.CreatedAt, &session.UpdatedAt, &session.IsActive); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (s *Store) GetSessionByID(ctx context.Context, id int) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx,
		`SELECT id, project_id, title, opencode_session_id, created_at, updated_at, is_active
		 FROM sessions WHERE id = $1`,
		id,
	).Scan(&session.ID, &session.ProjectID, &session.Title, &session.OpenCodeSessionID, &session.CreatedAt, &session.UpdatedAt, &session.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *Store) UpdateSessionTitle(ctx context.Context, id int, title string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET title = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		title, id,
	)
	return err
}

func (s *Store) UpdateSessionUpdatedAt(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
		id,
	)
	return err
}

func (s *Store) DeleteSession(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET is_active = FALSE WHERE id = $1`,
		id,
	)
	return err
}
```

**Step 6: Create `internal/store/message.go`**

```go
package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
)

type Message struct {
	ID        int
	SessionID int
	Role      string
	Content   string
	ToolCalls []ToolCall
	Status    string
	CreatedAt time.Time
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (s *Store) CreateMessage(ctx context.Context, sessionID int, role, content string, toolCalls []ToolCall, status string) (*Message, error) {
	tcJSON, _ := json.Marshal(toolCalls)
	var msg Message
	err := s.pool.QueryRow(ctx,
		`INSERT INTO messages (session_id, role, content, tool_calls, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, session_id, role, content, tool_calls, status, created_at`,
		sessionID, role, content, tcJSON, status,
	).Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &tcJSON, &msg.Status, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(tcJSON, &msg.ToolCalls)
	return &msg, nil
}

func (s *Store) GetMessagesBySession(ctx context.Context, sessionID int, limit, offset int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, role, content, tool_calls, status, created_at
		 FROM messages WHERE session_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		sessionID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var tcJSON []byte
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &tcJSON, &msg.Status, &msg.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(tcJSON, &msg.ToolCalls)
		messages = append(messages, msg)
	}
	return messages, nil
}

func (s *Store) SearchMessages(ctx context.Context, query string, limit int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, role, content, tool_calls, status, created_at
		 FROM messages
		 WHERE to_tsvector('simple', content) @@ plainto_tsquery('simple', $1)
		    OR content ILIKE '%' || $1 || '%'
		 ORDER BY created_at DESC LIMIT $2`,
		query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var tcJSON []byte
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &tcJSON, &msg.Status, &msg.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(tcJSON, &msg.ToolCalls)
		messages = append(messages, msg)
	}
	return messages, nil
}
```

---

## Phase 3: Authentication

### Task 3: JWT auth middleware and handlers

**Files:**
- Create: `internal/auth/jwt.go`
- Create: `internal/auth/password.go`
- Create: `internal/auth/middleware.go`
- Create: `internal/auth/auth_test.go`

**Step 1: Create `internal/auth/jwt.go`**

```go
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret     []byte
	expiry     time.Duration
	cookieName string
}

func NewJWTManager(secret string, expiry time.Duration, cookieName string) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		expiry:     expiry,
		cookieName: cookieName,
	}
}

func (m *JWTManager) Generate(userID int, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (m *JWTManager) CookieName() string {
	return m.cookieName
}
```

**Step 2: Create `internal/auth/password.go`**

```go
package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
```

**Step 3: Create `internal/auth/middleware.go`**

```go
package auth

import (
	"net/http"
	"strings"

	"github.com/ebachmann/opencode-webchat/internal/store"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	jwtManager *JWTManager
	store      *store.Store
}

func NewAuthMiddleware(jm *JWTManager, s *store.Store) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jm, store: s}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login and health endpoints
		if r.URL.Path == "/api/auth/login" || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		// Get token from cookie or Authorization header
		tokenStr := m.jwtManager.CookieName()
		if cookie, err := r.Cookie(tokenStr); err == nil && cookie.Value != "" {
			tokenStr = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		claims, err := m.jwtManager.Validate(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		user, err := m.store.GetUserByID(r.Context(), claims.UserID)
		if err != nil || user == nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		ctx := WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func GetUser(ctx context.Context) *store.User {
	if user, ok := ctx.Value(UserContextKey).(*store.User); ok {
		return user
	}
	return nil
}
```

---

## Phase 4: REST API Handlers

### Task 4: HTTP handlers for auth, projects, sessions, messages

**Files:**
- Create: `internal/http/handlers/auth.go`
- Create: `internal/http/handlers/projects.go`
- Create: `internal/http/handlers/sessions.go`
- Create: `internal/http/handlers/messages.go`
- Create: `internal/http/handlers/health.go`
- Create: `internal/http/router.go`

**Step 1: Create `internal/http/handlers/auth.go`**

```go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/store"
)

type AuthHandler struct {
	jwtManager *auth.JWTManager
	store      *store.Store
}

func NewAuthHandler(jm *auth.JWTManager, s *store.Store) *AuthHandler {
	return &AuthHandler{jwtManager: jm, store: s}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil || user == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtManager.Generate(user.ID, user.Username)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.jwtManager.CookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	json.NewEncoder(w).Encode(map[string]any{"username": user.Username})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.jwtManager.CookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"username": user.Username})
}
```

**Step 2: Create `internal/http/handlers/projects.go`**

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ebachmann/opencode-webchat/internal/store"
)

type ProjectsHandler struct {
	store *store.Store
}

func NewProjectsHandler(s *store.Store) *ProjectsHandler {
	return &ProjectsHandler{store: s}
}

type Project struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	RootPath   string `json:"root_path"`
	LastUsedAt string `json:"last_used_at"`
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	// For MVP, return a default project
	projects := []Project{
		{ID: 1, Name: "Default Project", RootPath: "/home/pi/projects"},
	}
	json.NewEncoder(w).Encode(projects)
}
```

**Step 3: Create `internal/http/handlers/sessions.go`**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/go-chi/chi/v5"
)

type SessionsHandler struct {
	store *store.Store
}

func NewSessionsHandler(s *store.Store) *SessionsHandler {
	return &SessionsHandler{store: s}
}

type CreateSessionRequest struct {
	ProjectID int    `json:"project_id"`
	Title     string `json:"title"`
}

func (h *SessionsHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, _ := strconv.Atoi(r.URL.Query().Get("project_id"))
	sessions, err := h.store.GetSessionsByProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, "failed to get sessions", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

func (h *SessionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		req.Title = "New Session"
	}

	session, err := h.store.CreateSession(r.Context(), req.ProjectID, req.Title)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *SessionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	session, err := h.store.GetSessionByID(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get session", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *SessionsHandler) UpdateTitle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateSessionTitle(r.Context(), id, req.Title); err != nil {
		http.Error(w, "failed to update session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.store.DeleteSession(r.Context(), id); err != nil {
		http.Error(w, "failed to delete session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

**Step 4: Create `internal/http/handlers/messages.go`**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/go-chi/chi/v5"
)

type MessagesHandler struct {
	store *store.Store
}

func NewMessagesHandler(s *store.Store) *MessagesHandler {
	return &MessagesHandler{store: s}
}

func (h *MessagesHandler) List(w http.ResponseWriter, r *http.Request) {
	sessionID, _ := strconv.Atoi(chi.URLParam(r, "sessionId"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 100
	}

	messages, err := h.store.GetMessagesBySession(r.Context(), sessionID, limit, offset)
	if err != nil {
		http.Error(w, "failed to get messages", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(messages)
}
```

**Step 5: Create `internal/http/handlers/health.go`**

```go
package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

**Step 6: Create `internal/http/router.go`**

```go
package http

import (
	"net/http"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/http/handlers"
	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(s *store.Store, jwtManager *auth.JWTManager) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Routes
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

	return r
}
```

---

## Phase 5: WebSocket & OpenCode PTY Manager

### Task 5: WebSocket hub and opencode process manager

**Files:**
- Create: `internal/ws/hub.go`
- Create: `internal/ws/client.go`
- Create: `internal/ws/messages.go`
- Create: `internal/opencode/manager.go`
- Create: `internal/opencode/process.go`

**Step 1: Create `internal/ws/messages.go`**

```go
package ws

// Message types for WebSocket protocol
type MessageType string

const (
	TypePrompt       MessageType = "prompt"
	TypeCancel       MessageType = "cancel"
	TypeApproveTool  MessageType = "approve_tool"
	TypeRejectTool   MessageType = "reject_tool"
	TypePing         MessageType = "ping"
	TypePong         MessageType = "pong"
	TypeToken        MessageType = "token"
	TypeToolCall     MessageType = "tool_call"
	TypeToolResult   MessageType = "tool_result"
	TypeDiffProposal MessageType = "diff_proposal"
	TypeDone         MessageType = "done"
	TypeError        MessageType = "error"
)

type InboundMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content,omitempty"`
	ID     string      `json:"id,omitempty"`
}

type OutboundMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content,omitempty"`
	ID     string      `json:"id,omitempty"`
	Data   any         `json:"data,omitempty"`
}
```

**Step 2: Create `internal/opencode/process.go`**

```go
package opencode

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

type Process struct {
	cmd    *exec.Cmd
	pty    *os.File
	stdin  io.Writer
	stdout *bufio.Reader
	stderr *bufio.Reader

	mu    sync.Mutex
	done  chan struct{}
}

func StartProcess(ctx context.Context, binaryPath string, args ...string) (*Process, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Env = os.Environ()

	ptyF, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start pty: %w", err)
	}

	return &Process{
		cmd:    cmd,
		pty:    ptyF,
		stdin:  ptyF,
		stdout: bufio.NewReader(ptyF),
		stderr: bufio.NewReader(ptyF),
		done:   make(chan struct{}),
	}, nil
}

func (p *Process) Write(ctx context.Context, data string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, err := p.stdin.Write([]byte(data))
	if err != nil {
		return err
	}
	_, err = p.stdin.Write([]byte("\n"))
	return err
}

func (p *Process) ReadLine(ctx context.Context) (string, error) {
	line, err := p.stdout.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return line, nil
		}
		return "", err
	}
	return line, nil
}

func (p *Process) Wait() error {
	err := p.cmd.Wait()
	close(p.done)
	return err
}

func (p *Process) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cmd.Process.Kill()
}

func (p *Process) IsDone() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}
```

**Step 3: Create `internal/opencode/manager.go`**

```go
package opencode

import (
	"context"
	"log"
	"sync"
)

type Manager struct {
	binaryPath string
	processes  map[int]*Process
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewManager(binaryPath string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		binaryPath: binaryPath,
		processes:  make(map[int]*Process),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (m *Manager) StartProcess(sessionID int) (*Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.processes[sessionID]; exists {
		return nil, ErrSessionExists
	}

	proc, err := StartProcess(m.ctx, m.binaryPath)
	if err != nil {
		return nil, err
	}

	m.processes[sessionID] = proc
	log.Printf("opencode process started for session %d", sessionID)
	return proc, nil
}

func (m *Manager) GetProcess(sessionID int) *Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processes[sessionID]
}

func (m *Manager) StopProcess(sessionID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	proc, exists := m.processes[sessionID]
	if !exists {
		return nil
	}

	if err := proc.Kill(); err != nil {
		log.Printf("error killing process for session %d: %v", sessionID, err)
	}

	delete(m.processes, sessionID)
	log.Printf("opencode process stopped for session %d", sessionID)
	return nil
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for sessionID, proc := range m.processes {
		proc.Kill()
		delete(m.processes, sessionID)
	}
	m.cancel()
}

var ErrSessionExists = context.DeadlineExceeded
```

**Step 4: Create `internal/ws/client.go`**

```go
package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait    = 10 * time.Second
	pongWait     = 60 * time.Second
	pingInterval = 30 * time.Second
	maxMessageSize = 512 * 1024
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	sessionID int
	userID    int
	send      chan []byte
	mu        sync.Mutex
}

func newClient(hub *Hub, conn *websocket.Conn, sessionID, userID int) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		sessionID: sessionID,
		userID:    userID,
		send:      make(chan []byte, 256),
	}
}

func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws error: %v", err)
			}
			break
		}

		var msg InboundMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}

		c.hub.handleMessage(c, &msg)
	}
}

func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.send:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mu.Unlock()
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.mu.Unlock()
				return
			}
			w.Write(msg)
			c.mu.Unlock()

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) sendJSON(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return ErrSendBufferFull
	}
}

var ErrSendBufferFull = context.DeadlineExceeded
```

**Step 5: Create `internal/ws/hub.go`**

```go
package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ebachmann/opencode-webchat/internal/opencode"
	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	opencodeMgr *opencode.Manager
	store       *store.Store

	clients    map[int]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage

	mu sync.RWMutex
}

type BroadcastMessage struct {
	SessionID int
	Data      []byte
	Exclude   *Client
}

func NewHub(om *opencode.Manager, s *store.Store) *Hub {
	return &Hub{
		opencodeMgr: om,
		store:       s,
		clients:     make(map[int]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.sessionID]; !ok {
				h.clients[client.sessionID] = make(map[*Client]bool)
			}
			h.clients[client.sessionID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.sessionID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.sessionID)
						h.opencodeMgr.StopProcess(client.sessionID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[msg.SessionID]
			h.mu.RUnlock()

			for client := range clients {
				if msg.Exclude != nil && client == msg.Exclude {
					continue
				}
				select {
				case client.send <- msg.Data:
				default:
					close(client.send)
					delete(clients, client)
				}
			}
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, sessionID, userID int) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	client := newClient(h, conn, sessionID, userID)
	h.register <- client

	go client.writePump(r.Context())
	go client.readPump(r.Context())
}

func (h *Hub) handleMessage(client *Client, msg *InboundMessage) {
	switch msg.Type {
	case TypePing:
		client.sendJSON(OutboundMessage{Type: TypePong})

	case TypePrompt:
		h.handlePrompt(client, msg)

	case TypeCancel:
		h.handleCancel(client)

	default:
		client.sendJSON(OutboundMessage{
			Type:    TypeError,
			Content: "unknown message type",
		})
	}
}

func (h *Hub) handlePrompt(client *Client, msg *InboundMessage) {
	proc := h.opencodeMgr.GetProcess(client.sessionID)
	if proc == nil {
		var err error
		proc, err = h.opencodeMgr.StartProcess(client.sessionID)
		if err != nil {
			client.sendJSON(OutboundMessage{
				Type:    TypeError,
				Content: "failed to start opencode process",
			})
			return
		}
		go h.streamOutput(client, proc)
	}

	h.store.CreateMessage(context.Background(), client.sessionID, "user", msg.Content, nil, "completed")

	if err := proc.Write(context.Background(), msg.Content); err != nil {
		client.sendJSON(OutboundMessage{
			Type:    TypeError,
			Content: "failed to write to opencode",
		})
	}
}

func (h *Hub) streamOutput(client *Client, proc *opencode.Process) {
	go func() {
		if err := proc.Wait(); err != nil {
			log.Printf("opencode process ended: %v", err)
		}
	}()

	for {
		line, err := proc.ReadLine(context.Background())
		if err != nil {
			break
		}

		outMsg := OutboundMessage{
			Type:    TypeToken,
			Content: line,
		}
		data, _ := json.Marshal(outMsg)

		h.broadcast <- &BroadcastMessage{
			SessionID: client.sessionID,
			Data:      data,
		}
	}

	doneMsg := OutboundMessage{Type: TypeDone}
	data, _ := json.Marshal(doneMsg)
	h.broadcast <- &BroadcastMessage{
		SessionID: client.sessionID,
		Data:      data,
	}
}

func (h *Hub) handleCancel(client *Client) {
	h.opencodeMgr.StopProcess(client.sessionID)
	client.sendJSON(OutboundMessage{
		Type:    TypeDone,
		Content: "cancelled",
	})
}
```

---

## Phase 6: Main Server & Embed

### Task 6: Wire everything together in main.go with embedded SPA

**Files:**
- Create: `cmd/server/main.go`
- Create: `web/.gitkeep`

**Step 1: Create `cmd/server/main.go`**

```go
package main

import (
	"context"
	"embed"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/config"
	"github.com/ebachmann/opencode-webchat/internal/http"
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
		log.Println("migrations not implemented — use goose CLI")
	}

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.CookieName)

	ocManager := opencode.NewManager(cfg.OpenCode.BinaryPath)
	defer ocManager.StopAll()

	hub := ws.NewHub(ocManager, s)
	hubCtx, hubCancel := context.WithCancel(context.Background())
	go hub.Run(hubCtx)

	router := http.NewRouter(s, jwtManager)

	wsHandler := &wsHandler{hub: hub, store: s, jwtManager: jwtManager}
	router.Mount("/ws", wsHandler)

	router.Mount("/", http.FileServer(embeddedFS))

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
	store      *store.Store
	jwtManager *auth.JWTManager
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
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

	sessionID := 1 // parse from sessionIDStr

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
```

**Step 2: Create `web/.gitkeep`**

```
# This directory will contain the Vue SPA build output
# The frontend will be served from here and embedded into the Go binary
```

---

## Verification Checklist

- [ ] `go build ./...` compiles without errors
- [ ] `go test ./...` passes
- [ ] `make test` runs full test suite
- [ ] Binary starts and serves on configured port
- [ ] WebSocket connection establishes and authenticates
- [ ] Health endpoint `/healthz` returns `{"status":"ok"}`
- [ ] Login creates JWT cookie
- [ ] Auth middleware protects API routes