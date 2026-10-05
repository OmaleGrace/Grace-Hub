package httpapi

import (
	"context"
	"net/http"
	"strings"

	"api/internal/users"
)

type ctxKey struct{}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		userID, err := users.ParseToken(s.secret, token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, userID)
		next(w, r.WithContext(ctx))
	}
}

func currentUserID(r *http.Request) int64 {
	id, _ := r.Context().Value(ctxKey{}).(int64)
	return id
}