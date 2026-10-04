package predictions

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Comparison struct {
	SystemAvg *float64
	UsersAvg  *float64
}

func ModelVsCommunity(ctx context.Context, pool *pgxpool.Pool, category string) (*Comparison, error) {
	const query = `
		SELECT
		  (SELECT avg(sp.score)::float8
		     FROM system_predictions sp JOIN events e ON e.id = sp.event_id
		    WHERE sp.score IS NOT NULL AND ($1 = '' OR e.category = $1)),
		  (SELECT avg(up.score)::float8
		     FROM user_predictions up JOIN events e ON e.id = up.event_id
		    WHERE up.score IS NOT NULL AND ($1 = '' OR e.category = $1))`

	var c Comparison
	if err := pool.QueryRow(ctx, query, category).Scan(&c.SystemAvg, &c.UsersAvg); err != nil {
		return nil, fmt.Errorf("compare model and community: %w", err)
	}
	return &c, nil
}