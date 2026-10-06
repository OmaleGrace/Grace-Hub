package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"api/internal/database"
	"api/internal/predictions"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: go run ./cmd/resolve <event id> <outcome>")
	}
	eventID, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Fatal("event id must be a number")
	}
	outcome := os.Args[2]

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

	n, err := predictions.Resolve(ctx, pool, eventID, outcome)
	if err != nil {
		log.Fatalf("resolve failed: %v", err)
	}
	log.Printf("event %d resolved as %q; %d user predictions scored", eventID, outcome, n)
}