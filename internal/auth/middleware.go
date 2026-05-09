package auth

import (
	"context"
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
		if r.URL.Path == "/api/auth/login" || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

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