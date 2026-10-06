package predictions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SystemPrediction struct {
	EventID       int64              `json:"event_id"`
	Predicted     string             `json:"predicted"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
	Odds          map[string]float64 `json:"odds"`
	ModelVersion  string             `json:"model_version"`
	Score         *float64           `json:"score"`
	CreatedAt     time.Time          `json:"created_at"`
}

func LatestSystemPrediction(ctx context.Context, pool *pgxpool.Pool, eventID int64) (*SystemPrediction, error) {
	const query = `
		SELECT event_id, predicted, confidence, probabilities, model_version, score, created_at
		FROM system_predictions
		WHERE event_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var p SystemPrediction
	var raw []byte
	err := pool.QueryRow(ctx, query, eventID).Scan(
		&p.EventID, &p.Predicted, &p.Confidence, &raw, &p.ModelVersion, &p.Score, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoSystemPrediction
	}
	if err != nil {
		return nil, fmt.Errorf("system prediction: %w", err)
	}

	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &p.Probabilities); err != nil {
			return nil, fmt.Errorf("decode probabilities: %w", err)
		}
		p.Odds = OddsFor(p.Probabilities)
	}
	return &p, nil
}