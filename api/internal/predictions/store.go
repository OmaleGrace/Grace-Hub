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
)

type UserPrediction struct {
	ID         int64
	UserID     int64
	EventID    int64
	Predicted  string
	Confidence float64
	Score      *float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func Submit(ctx context.Context, pool *pgxpool.Pool, userID, eventID int64,
	predicted string, confidence float64) (*UserPrediction, error) {

	predicted = strings.TrimSpace(predicted)
	if predicted == "" || confidence < 0 || confidence > 1 {
		return nil, ErrInvalidInput
	}

	const query = `
		INSERT INTO user_predictions (user_id, event_id, predicted, confidence)
		SELECT $1, e.id, $3, $4
		FROM events e
		WHERE e.id = $2
		  AND e.status = 'open'
		  AND e.opens_at <= now()
		  AND e.locks_at > now()
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
	return nil, diagnose(ctx, pool, eventID)
}

func diagnose(ctx context.Context, pool *pgxpool.Pool, eventID int64) error {
	const query = `
		SELECT (status = 'open' AND opens_at <= now() AND locks_at > now()),
		       EXISTS (SELECT 1 FROM system_predictions WHERE event_id = events.id)
		FROM events WHERE id = $1`

	var isOpen, hasSystem bool
	err := pool.QueryRow(ctx, query, eventID).Scan(&isOpen, &hasSystem)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEventNotFound
	}
	if err != nil {
		return fmt.Errorf("diagnose rejection: %w", err)
	}
	if !isOpen {
		return ErrEventNotOpen
	}
	if !hasSystem {
		return ErrNoSystemPrediction
	}
	return errors.New("prediction rejected for an unknown reason")
}