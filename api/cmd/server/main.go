package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/handler"
	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/ben-rieth/newsletter-api/internal/templates"
	"github.com/ben-rieth/newsletter-api/internal/users"
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

	rssService := feeds.NewRssService()
	emailService := newsletters.NewResendEmailService(&cfg)

	newsletterService := newsletters.NewNewsletterService(queries, pool)
	
	feeds.InitBlockedIPs()
	feedsService := feeds.NewFeedService(rssService, queries, pool)

	userService := users.NewUserService(queries, pool)
	
	tmpl, err := templates.ParseEmailTemplates()
	if err != nil {
		log.Fatalf("Could not get email templates: %v", err)
	}

	scheduler := newsletters.NewScheduler(
		newsletterService, 
		feedsService, 
		emailService,
		&newsletters.SchedulerConfig{
			MaxWorkers: 5,
		},
		tmpl,
	)
	if cfg.Environment != "dev" {
		go scheduler.KickOff(ctx)
	}

	mux := http.NewServeMux()
	humaConfig := huma.DefaultConfig("Newsletter API", "1.0.0")
	api := humago.New(mux, humaConfig)

	authHandler := handler.NewAuthHandler(*queries, cfg)
	authHandler.RegisterRoutes(api)

	if cfg.Environment == "dev" {
		scheudlerHandler := handler.NewSchedulerHandler(scheduler)
		scheudlerHandler.RegisterRoutes(api)
	}

	protectedApi := huma.NewGroup(api)
	protectedApi.UseMiddleware(auth.AuthMiddleware(api))

	newsletterHandler := handler.NewNewsletterHandler(queries, newsletterService)
	newsletterHandler.RegisterRoutes(protectedApi)

	feedsHandler := handler.NewFeedHandler(queries, feedsService)
	feedsHandler.RegisterRoutes(protectedApi)

	feedFilterHandler := handler.NewFeedFilterHandler(queries)
	feedFilterHandler.RegisterRoutes(protectedApi)

	userHandler := handler.NewUserHandler(queries, userService)
	userHandler.RegisterRoutes(protectedApi)

	exportHandler := handler.NewExportHandler(queries)
	exportHandler.RegisterRoutes(protectedApi)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{cfg.WebURL},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})

	log.Printf("Running on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), c.Handler(mux)))
}