package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/email"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/handler"
	"github.com/ben-rieth/newsletter-api/internal/jobs"
	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/ben-rieth/newsletter-api/internal/security"
	"github.com/ben-rieth/newsletter-api/internal/templates"
	"github.com/ben-rieth/newsletter-api/internal/ui"
	"github.com/ben-rieth/newsletter-api/internal/users"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const shutdownTimeout = 30 * time.Second

// Temporary: shows which header carries the visitor's IP once Cloudflare and
// Railway's edge have both rewritten the request. Remove after checking.
func logClientIPHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			slog.Info("client ip headers",
				"remoteAddr", r.RemoteAddr,
				"xForwardedFor", r.Header.Values("X-Forwarded-For"),
				"xRealIp", r.Header.Get("X-Real-Ip"),
				"cfConnectingIp", r.Header.Get("CF-Connecting-IP"),
			)
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load()
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	jobQueue := jobs.StartJobQueue(cfg)

	rssService := feeds.NewRssService()

	tmpl, err := templates.ParseEmailTemplates()
	if err != nil {
		log.Fatalf("Could not get email templates: %v", err)
	}
	emailService := email.NewResendEmailService(&cfg, tmpl, jobQueue)
	emailVerifyService := email.NewEmailVerifyService(ctx, queries, pool, cfg, emailService)

	newsletterService := newsletters.NewNewsletterService(queries, pool)
	issuesService := newsletters.NewIssuesService(queries, pool)

	feedsService := feeds.NewFeedService(rssService, queries, pool, jobQueue)

	userService := users.NewUserService(queries, pool)

	importService := newsletters.NewImportService(queries, pool, feedsService, jobQueue)
	if err := importService.ResumePendingImports(ctx); err != nil {
		log.Printf("Could not resume pending feed imports: %v", err)
	}

	scheduler := newsletters.NewScheduler(
		newsletterService,
		issuesService,
		feedsService,
		emailService,
		queries,
		&cfg,
		&newsletters.SchedulerConfig{
			MaxWorkers:        5,
			NewsletterTimeout: 300,
			TickTimeout:       600,
			SendRetryDelay:    2 * time.Second,
		},
	)

	go scheduler.KickOff(ctx)

	feeds.NewFailurePruner(queries, pool).KickOff(ctx)

	newsletters.NewIssuePruner(queries, pool).KickOff(ctx)

	apiMux := http.NewServeMux()

	humaConfig := huma.DefaultConfig("Newsletter API", "1.0.0")

	// The spec exists to feed `pnpm api:generate`, which runs against a dev server.
	// Served in production it is an unauthenticated map of every route and payload.
	if cfg.Environment != "dev" {
		humaConfig.OpenAPIPath = ""
		humaConfig.DocsPath = ""
		humaConfig.SchemasPath = ""
	}

	api := humago.New(apiMux, humaConfig)
	api.UseMiddleware(wideLog.WideLogMiddleware)

	authApi := huma.NewGroup(api)
	authApiRateLimiting := auth.NewRateLimitMiddleware(ctx, api, &cfg, 1, 5)
	authApi.UseMiddleware(authApiRateLimiting)

	signInLimiter := handler.NewSignInLimiter(ctx)
	authHandler := handler.NewAuthHandler(queries, pool, &cfg, emailVerifyService, userService, signInLimiter)
	authHandler.RegisterRoutes(authApi)

	rateLimiting := auth.NewRateLimitMiddleware(ctx, api, &cfg, 10, 30)

	publicApi := huma.NewGroup(api)
	publicApi.UseMiddleware(rateLimiting)

	unsubscribeHandler := handler.NewUnsubscribeHandler(queries, &cfg, emailService)
	unsubscribeHandler.RegisterRoutes(publicApi)

	linksHandler := handler.NewLinksHandler(queries, &cfg)
	linksHandler.RegisterRoutes(publicApi)

	protectedApi := huma.NewGroup(api)
	protectedApi.UseMiddleware(rateLimiting)
	protectedApi.UseMiddleware(auth.AuthMiddleware(api, &cfg, queries))

	newsletterHandler := handler.NewNewsletterHandler(queries, newsletterService)
	newsletterHandler.RegisterRoutes(protectedApi)

	if cfg.Environment == "dev" {
		debugApi := huma.NewGroup(api)
		debugApi.UseMiddleware(rateLimiting)
		schedulerHandler := handler.NewSchedulerHandler(scheduler, jobQueue)
		schedulerHandler.RegisterRoutes(debugApi)

		// Newsletter debug routes are user-scoped, so they need auth.
		protectedDebugApi := huma.NewGroup(api)
		protectedDebugApi.UseMiddleware(rateLimiting)
		protectedDebugApi.UseMiddleware(auth.AuthMiddleware(api, &cfg, queries))
		newsletterDebugHandler := handler.NewNewsletterDebugHandler(queries)
		newsletterDebugHandler.RegisterRoutes(protectedDebugApi)
	}

	feedsHandler := handler.NewFeedHandler(queries, feedsService)
	feedsHandler.RegisterRoutes(protectedApi)

	feedFilterHandler := handler.NewFeedFilterHandler(queries)
	feedFilterHandler.RegisterRoutes(protectedApi)

	userHandler := handler.NewUserHandler(queries, &cfg, userService, emailVerifyService, signInLimiter)
	userHandler.RegisterRoutes(protectedApi)

	exportHandler := handler.NewExportHandler(queries)
	exportHandler.RegisterRoutes(protectedApi)

	importHandler := handler.NewImportHandler(queries, importService)
	importHandler.RegisterRoutes(protectedApi)

	issuesHandler := handler.NewIssuesHandler(queries, issuesService)
	issuesHandler.RegisterRoutes(protectedApi)

	mux := http.NewServeMux()
	mux.Handle("/", ui.Handler())
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	csrf := http.NewCrossOriginProtection()

	// Headers is outermost, so a rejected request is answered with them too.
	handler := security.Headers(
		security.RequireOriginSecret(logClientIPHeaders(csrf.Handler(mux)), cfg.OriginAuthSecret),
		cfg.Environment == "prod",
	)

	// Without these a client that trickles bytes holds its connection open
	// forever. WriteTimeout bounds whole handlers, so it has to clear the slowest
	// one: a feed fetch, capped at about 15s by the feeds HTTP client.
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	// Jobs outlive the request that queued them, so the queue drains after the
	// server stops accepting rather than alongside it.
	if err := jobQueue.Shutdown(shutdownCtx); err != nil {
		log.Printf("job queue shutdown error: %v", err)
	}
}
