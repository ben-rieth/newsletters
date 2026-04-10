package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/handler"
	"github.com/ben-rieth/newsletter-api/internal/scheduler"
	"github.com/ben-rieth/newsletter-api/internal/service"
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

	newsletterService := service.NewNewsletterService(queries)
	
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	rssService := service.NewRssService(httpClient)
	
	scheduler := scheduler.NewScheduler(
		newsletterService, 
		rssService, 
		&scheduler.SchedulerConfig{
			MaxWorkers: 5,
		},
	)
	go scheduler.KickOff(ctx)

	mux := http.NewServeMux()
	humaConfig := huma.DefaultConfig("Newsletter API", "1.0.0")
	api := humago.New(mux, humaConfig)

	authHandler := handler.NewUserHandler(*queries, cfg)
	authHandler.RegisterRoutes(api)

	protectedApi := huma.NewGroup(api)
	protectedApi.UseMiddleware(auth.AuthMiddleware(api))

	newsletterHandler := handler.NewNewsletterHandler(queries, pool)
	newsletterHandler.RegisterRoutes(protectedApi)

	feedsHandler := handler.NewFeedHandler(queries)
	feedsHandler.RegisterRoutes(protectedApi)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{cfg.WebURL},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})

	log.Printf("Running on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), c.Handler(mux)))
}