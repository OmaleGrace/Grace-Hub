package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	UnitsPerPoint = 100
	SignupBonus   = 1000 * UnitsPerPoint
)

type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Entry struct {
	ID        int64     `json:"id"`
	Amount    int64     `json:"amount_units"`
	Reason    string    `json:"reason"`
	EventID   *int64    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

func GrantSignupBonus(ctx context.Context, q Execer, userID int64) error {
	_, err := q.Exec(ctx,
		`INSERT INTO wallet_entries (user_id, amount, reason) VALUES ($1, $2, 'signup_bonus')`,
		userID, SignupBonus)
	if err != nil {
		return fmt.Errorf("grant signup bonus: %w", err)
	}
	return nil
}

func Balance(ctx context.Context, pool *pgxpool.Pool, userID int64) (int64, error) {
	var total int64
	err := pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0)::bigint FROM wallet_entries
		 WHERE user_id = $1 AND currency = 'points'`, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("balance: %w", err)
	}
	return total, nil
}

func History(ctx context.Context, pool *pgxpool.Pool, userID int64, limit int) ([]Entry, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, amount, reason, event_id, created_at
		FROM wallet_entries
		WHERE user_id = $1 AND currency = 'points'
		ORDER BY id DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("history: %w", err)
	}
	defer rows.Close()

	entries := []Entry{}
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Amount, &e.Reason, &e.EventID, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read history: %w", err)
	}
	return entries, nil
}

func AddEntry(ctx context.Context, q Execer, userID, amount int64, reason string, eventID, betID int64) error {
	_, err := q.Exec(ctx, `
		INSERT INTO wallet_entries (user_id, amount, reason, event_id, bet_id)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, amount, reason, eventID, betID)
	if err != nil {
		return fmt.Errorf("add wallet entry: %w", err)
	}
	return nil
}