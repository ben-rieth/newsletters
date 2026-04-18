package config

import (
	"log"
	"net/mail"
	"os"
)

type Config struct {
	DatabaseURL string
	Port string
	Host string
	WebURL string
	JWTSecret string
	ResendAPIKey string
	Environment string
	NewsletterSenderEmail mail.Address
}

func Load() Config {
	cfg := Config {
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port: os.Getenv("PORT"),
		Host: os.Getenv("HOST"),
		WebURL: os.Getenv("WEB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		Environment: os.Getenv("ENVIRONMENT"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if cfg.WebURL == "" {
		log.Fatal("WEB_URL is required")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	if cfg.ResendAPIKey == "" {
		log.Fatal("RESEND_API_KEY")
	}

	if !(cfg.Environment == "dev" || cfg.Environment == "prod") {
		log.Fatalf("Invalid value for environment: %s", cfg.Environment)
	}

	nlSenderEmail := os.Getenv("NEWSLETTER_SENDER_EMAIL")
	if nlSenderEmail == "" {
		log.Fatalf("NEWSLETTER_SENDER_EMAIL is required")
	}

	addr, err := mail.ParseAddress(nlSenderEmail)
	if err != nil {
		log.Fatalf("NEWSLETTER_SENDER_EMAIL is invalid: %v", err)
	}
	cfg.NewsletterSenderEmail = *addr

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}

	return cfg
}