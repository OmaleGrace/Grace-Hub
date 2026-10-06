package predictions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAlreadyResolved = errors.New("event is already resolved")
	ErrInvalidOutcome  = errors.New("outcome is not one of the event's options")
	ErrNotResolvable   = errors.New("event cannot be resolved yet")
)

func Resolve(ctx context.Context, pool *pgxpool.Pool, eventID int64, outcome string) (int64, error) {
	outcome = strings.ToLower(strings.TrimSpace(outcome))
	if outcome == "" {
		return 0, ErrInvalidInput
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int64
	err = tx.QueryRow(ctx, `
		UPDATE events
		SET status = 'resolved', outcome = $2, resolved_at = now()
		WHERE id = $1 AND status <> 'resolved' AND locks_at <= now() AND $2::text = ANY(options)
		RETURNING id`, eventID, outcome).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, diagnoseResolve(ctx, pool, eventID, outcome)
	}
	if err != nil {
		return 0, fmt.Errorf("resolve event: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE user_predictions
		SET score = 1 - power(
			confidence - (CASE WHEN lower(trim(predicted)) = $2 THEN 1 ELSE 0 END), 2)
		WHERE event_id = $1`, eventID, outcome)
	if err != nil {
		return 0, fmt.Errorf("score predictions: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE system_predictions
		SET score = 1 - power(
			confidence - (CASE WHEN lower(trim(predicted)) = $2 THEN 1 ELSE 0 END), 2)
		WHERE event_id = $1`, eventID, outcome)
	if err != nil {
		return 0, fmt.Errorf("score system predictions: %w", err)
	}

	if err := settleBets(ctx, tx, eventID, outcome); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return tag.RowsAffected(), nil
}

func diagnoseResolve(ctx context.Context, pool *pgxpool.Pool, eventID int64, outcome string) error {
	var status string
	var lockPassed, validOutcome bool
	err := pool.QueryRow(ctx,
		`SELECT status, locks_at <= now(), $2::text = ANY(options) FROM events WHERE id = $1`,
		eventID, outcome,
	).Scan(&status, &lockPassed, &validOutcome)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEventNotFound
	}
	if err != nil {
		return fmt.Errorf("diagnose resolve: %w", err)
	}
	if status == "resolved" {
		return ErrAlreadyResolved
	}
	if !lockPassed {
		return ErrNotResolvable
	}
	if !validOutcome {
		return ErrInvalidOutcome
	}
	return errors.New("event could not be resolved")
}

func settleBets(ctx context.Context, tx pgx.Tx, eventID int64, outcome string) error {
	if _, err := tx.Exec(ctx, `
		UPDATE bets
		SET status = 'won', payout = floor(stake * odds)::bigint, settled_at = now()
		WHERE event_id = $1 AND status = 'open' AND choice = $2`,
		eventID, outcome); err != nil {
		return fmt.Errorf("mark winning bets: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE bets
		SET status = 'lost', payout = 0, settled_at = now()
		WHERE event_id = $1 AND status = 'open' AND choice <> $2`,
		eventID, outcome); err != nil {
		return fmt.Errorf("mark losing bets: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO wallet_entries (user_id, amount, reason, event_id, bet_id)
		SELECT user_id, payout, 'payout', event_id, id
		FROM bets
		WHERE event_id = $1 AND status = 'won'`,
		eventID); err != nil {
		return fmt.Errorf("pay winners: %w", err)
	}
	return nil
}