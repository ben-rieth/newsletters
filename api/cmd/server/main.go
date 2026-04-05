package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/handler"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	godotenv.Load()
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	mux := http.NewServeMux()
	humaConfig := huma.DefaultConfig("Newsletter API", "1.0.0")
	api := humago.New(mux, humaConfig)

	newsletterHandler := handler.NewNewsletterHandler(queries, pool)
	newsletterHandler.RegisterRoutes(api)

	feedsHandler := handler.NewFeedHandler(queries)
	feedsHandler.RegisterRoutes(api)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{cfg.WebURL},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})

	log.Printf("Running on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), c.Handler(mux)))
}