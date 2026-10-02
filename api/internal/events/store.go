package events

import (
	"context"
	"fmt"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"

)

func Create(ctx context.Context, pool *pgxpool.Pool, e *Event) error {
	if e.Details == nil {
		e.Details = []byte("{}")
	}

	const query = `
		INSERT INTO events (category, title, details, opens_at, locks_at, resolves_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, status, created_at`

	err := pool.QueryRow(ctx, query,
		e.Category, e.Title, e.Details, e.OpensAt, e.LocksAt, e.ResolvesAt,
	).Scan(&e.ID, &e.Status, &e.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

var ErrNotFound = errors.New("event not found")

const selectColumns = `id, category, title, details, opens_at, locks_at,
	resolves_at, status, outcome, resolved_at, created_at`

func scanEvent(row pgx.Row) (*Event, error) {
	var e Event
	err := row.Scan(&e.ID, &e.Category, &e.Title, &e.Details, &e.OpensAt,
		&e.LocksAt, &e.ResolvesAt, &e.Status, &e.Outcome, &e.ResolvedAt, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func Get(ctx context.Context, pool *pgxpool.Pool, id int64) (*Event, error) {
	row := pool.QueryRow(ctx,
		`SELECT `+selectColumns+` FROM events WHERE id = $1`, id)

	e, err := scanEvent(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	return e, nil
}

func List(ctx context.Context, pool *pgxpool.Pool, category, status string) ([]Event, error) {
	const query = `SELECT ` + selectColumns + `
		FROM events
		WHERE ($1 = '' OR category = $1)
		  AND ($2 = '' OR status = $2)
		ORDER BY opens_at DESC`

	rows, err := pool.Query(ctx, query, category, status)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var result []Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		result = append(result, *e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}
	return result, nil
}

func LockExpired(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	tag, err := pool.Exec(ctx,
		`UPDATE events SET status = 'locked'
		 WHERE status = 'open' AND locks_at <= now()`)
	if err != nil {
		return 0, fmt.Errorf("lock expired events: %w", err)
	}
	return tag.RowsAffected(), nil
}