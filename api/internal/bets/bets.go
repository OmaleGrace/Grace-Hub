package bets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api/internal/predictions"
	"api/internal/wallet"
)

const (
	MinStakePoints = 1
	MaxStakePoints = 500
)

var (
	ErrInvalidInput      = errors.New("invalid bet input")
	ErrEventNotFound     = errors.New("event not found")
	ErrEventNotOpen      = errors.New("event is not open for bets")
	ErrInvalidOption     = errors.New("choice is not one of the event's options")
	ErrNoOdds            = errors.New("odds are not available for this event")
	ErrInsufficientFunds = errors.New("not enough points")
	ErrAlreadyBet        = errors.New("you already placed a bet on this event")
)

type Bet struct {
	ID          int64     `json:"id"`
	EventID     int64     `json:"event_id"`
	Choice      string    `json:"choice"`
	StakeUnits  int64     `json:"stake_units"`
	Odds        float64   `json:"odds"`
	Status      string    `json:"status"`
	PayoutUnits *int64    `json:"payout_units"`
	CreatedAt   time.Time `json:"created_at"`
}

func Place(ctx context.Context, pool *pgxpool.Pool, userID, eventID int64,
	choice string, stakePoints int64) (*Bet, error) {

	choice = strings.ToLower(strings.TrimSpace(choice))
	if choice == "" || stakePoints < MinStakePoints || stakePoints > MaxStakePoints {
		return nil, ErrInvalidInput
	}
	stake := stakePoints * wallet.UnitsPerPoint

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var lockedID int64
	if err := tx.QueryRow(ctx,
		`SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&lockedID); err != nil {
		return nil, fmt.Errorf("lock user: %w", err)
	}

	var open bool
	var options []string
	var raw []byte
	err = tx.QueryRow(ctx, `
		SELECT (e.status = 'open' AND e.opens_at <= clock_timestamp()
		        AND e.locks_at > clock_timestamp()),
		       e.options,
		       (SELECT sp.probabilities FROM system_predictions sp
		         WHERE sp.event_id = e.id ORDER BY sp.created_at DESC LIMIT 1)
		FROM events e WHERE e.id = $1`, eventID).Scan(&open, &options, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEventNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load event: %w", err)
	}
	if !open {
		return nil, ErrEventNotOpen
	}

	validChoice := false
	for _, o := range options {
		if o == choice {
			validChoice = true
		}
	}
	if !validChoice {
		return nil, ErrInvalidOption
	}
	if len(raw) == 0 {
		return nil, ErrNoOdds
	}
	var probs map[string]float64
	if err := json.Unmarshal(raw, &probs); err != nil {
		return nil, fmt.Errorf("decode probabilities: %w", err)
	}
	odds, ok := predictions.OddsFor(probs)[choice]
	if !ok {
		return nil, ErrNoOdds
	}

	var balance int64
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)::bigint FROM wallet_entries
		WHERE user_id = $1 AND currency = 'points'`, userID).Scan(&balance); err != nil {
		return nil, fmt.Errorf("balance: %w", err)
	}
	if balance < stake {
		return nil, ErrInsufficientFunds
	}

	var b Bet
	err = tx.QueryRow(ctx, `
		INSERT INTO bets (user_id, event_id, choice, stake, odds)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, event_id, choice, stake, odds::float8, status, payout, created_at`,
		userID, eventID, choice, stake, odds,
	).Scan(&b.ID, &b.EventID, &b.Choice, &b.StakeUnits, &b.Odds, &b.Status, &b.PayoutUnits, &b.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, ErrAlreadyBet
	}
	if err != nil {
		return nil, fmt.Errorf("insert bet: %w", err)
	}

	if err := wallet.AddEntry(ctx, tx, userID, -stake, "stake", eventID, b.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &b, nil
}