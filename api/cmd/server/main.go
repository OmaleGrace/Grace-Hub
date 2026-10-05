package main

import (
	"context"
	"crypto/rand"
	"log"
	"net/http"
	"os"
	"time"

	"api/internal/database"
	"api/internal/events"
	"api/internal/httpapi"
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

	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			log.Fatal(err)
		}
		log.Print("JWT_SECRET not set: using a random secret; logins will not survive a restart")
	}

	go events.RunLocker(ctx, pool, 30*time.Second)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.New(pool, secret),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}