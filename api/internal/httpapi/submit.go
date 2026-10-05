package httpapi

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"api/internal/predictions"
)

func (s *Server) submitPrediction(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "event id must be a number")
		return
	}

	var in struct {
		Predicted  string  `json:"predicted"`
		Confidence float64 `json:"confidence"`
	}
	if !decode(w, r, &in) {
		return
	}

	p, err := predictions.Submit(r.Context(), s.pool, currentUserID(r), eventID, in.Predicted, in.Confidence)
	switch {
	case errors.Is(err, predictions.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "predicted must not be empty and confidence must be between 0 and 1")
	case errors.Is(err, predictions.ErrInvalidOption):
		writeError(w, http.StatusBadRequest, "predicted must be one of the event's options")
	case errors.Is(err, predictions.ErrEventNotFound):
		writeError(w, http.StatusNotFound, "event not found")
	case errors.Is(err, predictions.ErrEventNotOpen):
		writeError(w, http.StatusConflict, "event is not open for predictions")
	case errors.Is(err, predictions.ErrNoSystemPrediction):
		writeError(w, http.StatusConflict, "predictions are not open yet for this event")
	case err != nil:
		log.Printf("submit prediction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, p)
	}
}