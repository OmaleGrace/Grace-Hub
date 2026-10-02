package events

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunLocker(ctx context.Context, pool *pgxpool.Pool, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := LockExpired(ctx, pool)
			if err != nil {
				log.Printf("locker error: %v", err)
				continue
			}
			if n > 0 {
				log.Printf("locked %d events", n)
			}
		}
	}
}