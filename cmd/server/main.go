package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authdelivery "github.com/marcello/saas-poc/internal/delivery/auth"
	"github.com/marcello/saas-poc/internal/delivery/dashboard"
	"github.com/marcello/saas-poc/internal/delivery/system"
	"github.com/marcello/saas-poc/internal/delivery/website"
	"github.com/marcello/saas-poc/internal/middleware"
	"github.com/marcello/saas-poc/internal/todos"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// preShutdownDelay keeps serving after SIGTERM while load balancers drop this
// instance from their pool; preShutdownDelay + shutdownTimeout must stay below
// the orchestrator grace period (Kubernetes terminationGracePeriodSeconds).
const (
	preShutdownDelay = 5 * time.Second
	shutdownTimeout  = 20 * time.Second
)

func main() {
	signalContext, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://saas_poc:saas_poc@localhost:5432/saas_poc?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("configure PostgreSQL pool: %v", err)
	}
	defer pool.Close()

	pingContext, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := pool.Ping(pingContext); err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}

	router := chi.NewRouter()

	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(middleware.Logging)

	authConfig, err := authdelivery.NewConfig(
		environmentOrDefault("ZITADEL_ISSUER", "http://auth.localhost:8081"),
		os.Getenv("ZITADEL_CLIENT_ID"),
		environmentOrDefault("APPLICATION_ORIGIN", "http://localhost:8080"),
	)
	if err != nil {
		log.Fatalf("configure authentication: %v", err)
	}
	discoveryContext, cancelDiscovery := context.WithTimeout(context.Background(), 10*time.Second)
	tokenVerifier, err := middleware.NewOIDCAccessTokenVerifier(discoveryContext, authConfig.Issuer, authConfig.ClientID)
	cancelDiscovery()
	if err != nil {
		log.Fatalf("configure token verifier: %v", err)
	}

	website.RegisterRoutes(router)
	authdelivery.RegisterRoutes(router, authConfig)
	todoRepository := todos.NewPostgresTodoRepository(pool)
	todoService := todos.NewTodoService(todoRepository)
	dashboard.RegisterRoutes(router, todoService, middleware.RequireBearer(tokenVerifier, "super-admin", "regular-user"))
	system.RegisterRoutes(router)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	server.RegisterOnShutdown(dashboard.Shutdown)

	fmt.Println(`
╔════════════════════════════════════════════════════════╗
║       SaaS Gestionale PoC - Minimal Stack 2026        ║
╚════════════════════════════════════════════════════════╝

📡 Server running on http://localhost:8080

📊 Routes:
	GET  /                → Website landing (Full HTML page with Templ)
	GET  /dashboard       → Dashboard (Full HTML page with Templ)
	GET  /dashboard/events → Server-Sent Events demo (stream + broadcast button)
	GET  /dashboard/loan  → Loan application demo (Reactive HTMX validation page)
	GET/POST/PUT/PATCH/DELETE /dashboard/todos → PostgreSQL-backed TODO list
	GET  /dashboard/status → Status component (HTMX endpoint, returns HTML fragment)
	POST /dashboard/loan/validate → Loan form cross-field validation (HTML fragment)
	POST /dashboard/loan/submit  → Loan form submission simulation (HTML fragment)
   GET  /api/info        → API Info (JSON REST endpoint)
   GET  /health          → Health check (JSON)

🛠️  Tech Stack:
   • Router: github.com/go-chi/chi/v5
   • Templates: github.com/a-h/templ (type-safe HTML)
   • Frontend: HTMX (CDN) + Bootstrap styling
	• Database: PostgreSQL 18 via github.com/jackc/pgx/v5
	• Migrations: github.com/pressly/goose
   • REST API: encoding/json (stdlib)
   • Middleware: Chi + custom handlers

🔐 Security & Compliance:
	• Development database defaults can be overridden with DATABASE_URL
	• Parameterized SQL queries through pgx
   • Type-safe HTML rendering (Templ prevents XSS)
   • Standard library for JSON encoding

💡 Architecture:
	• HTMX talks to /dashboard/status → server renders partial HTML → browser updates DOM
	• TODO handlers call a service layer backed by a PostgreSQL repository
   • Classic REST calls to /api/info → server returns JSON with stdlib
	• Website landing and dashboard use separate Templ layouts

Press Ctrl+C to stop the server.
	`)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve HTTP: %v", err)
		}
	case <-signalContext.Done():
		stopSignals()
		log.Printf("shutdown signal received, waiting %s before draining connections", preShutdownDelay)
		time.Sleep(preShutdownDelay)
		log.Println("draining connections")

		shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("graceful shutdown incomplete: %v", err)
			if err := server.Close(); err != nil {
				log.Printf("force close listeners: %v", err)
			}
		}

		if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve HTTP: %v", err)
		}

		log.Println("shutdown complete")
	}
}

func environmentOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
