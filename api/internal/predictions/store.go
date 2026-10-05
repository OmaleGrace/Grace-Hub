package predictions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventNotOpen       = errors.New("event is not open for predictions")
	ErrNoSystemPrediction = errors.New("system has not predicted this event yet")
	ErrInvalidInput       = errors.New("invalid prediction input")
	ErrInvalidOption      = errors.New("prediction is not one of the event's options")
)

type UserPrediction struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	EventID    int64     `json:"event_id"`
	Predicted  string    `json:"predicted"`
	Confidence float64   `json:"confidence"`
	Score      *float64  `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func Submit(ctx context.Context, pool *pgxpool.Pool, userID, eventID int64,
	predicted string, confidence float64) (*UserPrediction, error) {

	predicted = strings.ToLower(strings.TrimSpace(predicted))
	if predicted == "" || confidence < 0 || confidence > 1 {
		return nil, ErrInvalidInput
	}

	const query = `
		INSERT INTO user_predictions (user_id, event_id, predicted, confidence)
		SELECT $1, e.id, $3::text, $4
		FROM events e
		WHERE e.id = $2
		  AND e.status = 'open'
		  AND e.opens_at <= now()
		  AND e.locks_at > now()
		  AND $3::text = ANY(e.options)
		  AND EXISTS (SELECT 1 FROM system_predictions sp WHERE sp.event_id = e.id)
		ON CONFLICT (user_id, event_id) DO UPDATE
		  SET predicted = EXCLUDED.predicted,
		      confidence = EXCLUDED.confidence,
		      updated_at = now()
		RETURNING id, user_id, event_id, predicted, confidence, score, created_at, updated_at`

	var p UserPrediction
	err := pool.QueryRow(ctx, query, userID, eventID, predicted, confidence).Scan(
		&p.ID, &p.UserID, &p.EventID, &p.Predicted, &p.Confidence,
		&p.Score, &p.CreatedAt, &p.UpdatedAt)
	if err == nil {
		return &p, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("submit prediction: %w", err)
	}
	return nil, diagnose(ctx, pool, eventID, predicted)
}

func diagnose(ctx context.Context, pool *pgxpool.Pool, eventID int64, predicted string) error {
	const query = `
		SELECT (status = 'open' AND opens_at <= now() AND locks_at > now()),
		       $2::text = ANY(options),
		       EXISTS (SELECT 1 FROM system_predictions WHERE event_id = events.id)
		FROM events WHERE id = $1`

	var isOpen, validOption, hasSystem bool
	err := pool.QueryRow(ctx, query, eventID, predicted).Scan(&isOpen, &validOption, &hasSystem)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEventNotFound
	}
	if err != nil {
		return fmt.Errorf("diagnose rejection: %w", err)
	}
	if !isOpen {
		return ErrEventNotOpen
	}
	if !validOption {
		return ErrInvalidOption
	}
	if !hasSystem {
		return ErrNoSystemPrediction
	}
	return errors.New("prediction rejected for an unknown reason")
}