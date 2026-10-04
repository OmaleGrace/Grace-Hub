package predictions

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LeaderboardEntry struct {
	Rank        int
	UserID      int64
	Username    string
	Predictions int64
	AvgScore    float64
}

func Leaderboard(ctx context.Context, pool *pgxpool.Pool, category string,
	minPredictions, limit int) ([]LeaderboardEntry, error) {

	const query = `
		SELECT u.id, u.username, count(*) AS n, avg(up.score)::float8 AS avg_score
		FROM user_predictions up
		JOIN users u ON u.id = up.user_id
		JOIN events e ON e.id = up.event_id
		WHERE up.score IS NOT NULL
		  AND ($1 = '' OR e.category = $1)
		GROUP BY u.id, u.username
		HAVING count(*) >= $2
		ORDER BY avg_score DESC, n DESC, u.username
		LIMIT $3`

	rows, err := pool.Query(ctx, query, category, minPredictions, limit)
	if err != nil {
		return nil, fmt.Errorf("leaderboard: %w", err)
	}
	defer rows.Close()

	var result []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.Predictions, &e.AvgScore); err != nil {
			return nil, fmt.Errorf("scan leaderboard: %w", err)
		}
		e.Rank = len(result) + 1
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read leaderboard: %w", err)
	}
	return result, nil
}