package httpapi

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"api/internal/bets"
	"api/internal/wallet"
)

func (s *Server) placeBet(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "event id must be a number")
		return
	}

	var in struct {
		Choice      string `json:"choice"`
		StakePoints int64  `json:"stake_points"`
	}
	if !decode(w, r, &in) {
		return
	}

	userID := currentUserID(r)
	b, err := bets.Place(r.Context(), s.pool, userID, eventID, in.Choice, in.StakePoints)
	switch {
	case errors.Is(err, bets.ErrInvalidInput):
		writeError(w, http.StatusBadRequest,
			"choice is required and stake_points must be a whole number from 1 to 500")
	case errors.Is(err, bets.ErrInvalidOption):
		writeError(w, http.StatusBadRequest, "choice must be one of the event's options")
	case errors.Is(err, bets.ErrEventNotFound):
		writeError(w, http.StatusNotFound, "event not found")
	case errors.Is(err, bets.ErrEventNotOpen):
		writeError(w, http.StatusConflict, "event is not open for bets")
	case errors.Is(err, bets.ErrNoOdds):
		writeError(w, http.StatusConflict, "betting is not open yet for this event")
	case errors.Is(err, bets.ErrInsufficientFunds):
		writeError(w, http.StatusConflict, "not enough points")
	case errors.Is(err, bets.ErrAlreadyBet):
		writeError(w, http.StatusConflict, "you already placed a bet on this event")
	case err != nil:
		log.Printf("place bet: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		balance, berr := wallet.Balance(r.Context(), s.pool, userID)
		if berr != nil {
			log.Printf("balance after bet: %v", berr)
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"bet":           b,
			"balance_units": balance,
		})
	}
}