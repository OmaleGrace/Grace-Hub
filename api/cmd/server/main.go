package main

import (
	"context"
	"log"
	"os"

	"api/internal/database"
	"api/internal/predictions"
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

	c, err := predictions.ModelVsCommunity(ctx, pool, "")
	if err != nil {
		log.Fatal(err)
	}
	show := func(label string, v *float64) {
		if v == nil {
			log.Printf("%s: no scored predictions yet", label)
			return
		}
		log.Printf("%s: average score %.4f", label, *v)
	}
	show("system   ", c.SystemAvg)
	show("community", c.UsersAvg)
}