package main

import (
	"context"
	"log"
	"os"
	"time"

	"api/internal/database"
	"api/internal/events"
)

func main() {
	ctx := context.Background()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://predict:predict_dev_password@localhost:5432/predictions"
	}

	pool, err := database.Connect(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	now := time.Now()
	e := &events.Event{
		Category:   "sports",
		Title:      "Already past its lock time",
		OpensAt:    now.Add(-2 * time.Hour),
		LocksAt:    now.Add(-1 * time.Hour),
		ResolvesAt: now.Add(1 * time.Hour),
	}
	if err := events.Create(ctx, pool, e); err != nil {
		log.Fatal(err)
	}
	log.Printf("created id=%d status=%s", e.ID, e.Status)

	n, err := events.LockExpired(ctx, pool)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("locked %d events", n)

	got, err := events.Get(ctx, pool, e.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("after locking: id=%d status=%s", got.ID, got.Status)
}