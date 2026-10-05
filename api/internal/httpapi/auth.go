package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"api/internal/users"
)

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}

	u, err := users.Register(r.Context(), s.pool, in.Username, in.Email, in.Password)
	switch {
	case errors.Is(err, users.ErrInvalidInput):
		writeError(w, http.StatusBadRequest,
			"username must be 3-30 letters, numbers or underscores; email must be valid; password must be 8-72 characters")
	case errors.Is(err, users.ErrTaken):
		writeError(w, http.StatusConflict, "username or email already in use")
	case err != nil:
		log.Printf("register: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusCreated, u)
	}
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}

	u, err := users.Authenticate(r.Context(), s.pool, in.Login, in.Password)
	if errors.Is(err, users.ErrBadCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	if err != nil {
		log.Printf("login: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	token, err := users.IssueToken(s.secret, u.ID, 24*time.Hour)
	if err != nil {
		log.Printf("issue token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}