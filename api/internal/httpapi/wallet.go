package httpapi

import (
	"log"
	"net/http"

	"api/internal/wallet"
)

func (s *Server) myWallet(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)

	balance, err := wallet.Balance(r.Context(), s.pool, userID)
	if err != nil {
		log.Printf("wallet balance: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	history, err := wallet.History(r.Context(), s.pool, userID, 20)
	if err != nil {
		log.Printf("wallet history: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"balance_units":   balance,
		"units_per_point": wallet.UnitsPerPoint,
		"history":         history,
	})
}