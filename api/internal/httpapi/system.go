package httpapi

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"api/internal/predictions"
)

func (s *Server) systemPrediction(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "event id must be a number")
		return
	}

	p, err := predictions.LatestSystemPrediction(r.Context(), s.pool, id)
	if errors.Is(err, predictions.ErrNoSystemPrediction) {
		writeError(w, http.StatusNotFound, "no system prediction for this event yet")
		return
	}
	if err != nil {
		log.Printf("system prediction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, p)
}