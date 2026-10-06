package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"api/internal/events"
	"api/internal/predictions"
)

type Server struct {
	pool   *pgxpool.Pool
	secret []byte
}

func New(pool *pgxpool.Pool, secret []byte) http.Handler {
	s := &Server{pool: pool, secret: secret}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /events", s.listEvents)
	mux.HandleFunc("GET /events/{id}", s.getEvent)
	mux.HandleFunc("GET /leaderboard", s.leaderboard)
	mux.HandleFunc("POST /register", s.register)
	mux.HandleFunc("POST /login", s.login)
	mux.HandleFunc("POST /events/{id}/predictions", s.requireAuth(s.submitPrediction))
	mux.Handle("GET /", staticHandler())
	mux.HandleFunc("GET /events/{id}/system-prediction", s.systemPrediction)
	mux.HandleFunc("GET /login", pageHandler("login.html"))
	mux.HandleFunc("GET /signup", pageHandler("signup.html"))
	mux.HandleFunc("GET /me/wallet", s.requireAuth(s.myWallet))
	mux.HandleFunc("POST /events/{id}/bets", s.requireAuth(s.placeBet))
	return mux
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func intParam(r *http.Request, name string, def, max int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || n < 1 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := events.List(r.Context(), s.pool, q.Get("category"), q.Get("status"))
	if err != nil {
		log.Printf("list events: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if list == nil {
		list = []events.Event{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "event id must be a number")
		return
	}
	e, err := events.Get(r.Context(), s.pool, id)
	if errors.Is(err, events.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		log.Printf("get event: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	minPredictions := intParam(r, "min", 1, 1000)
	limit := intParam(r, "limit", 10, 100)

	board, err := predictions.Leaderboard(r.Context(), s.pool, category, minPredictions, limit)
	if err != nil {
		log.Printf("leaderboard: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if board == nil {
		board = []predictions.LeaderboardEntry{}
	}
	writeJSON(w, http.StatusOK, board)
}