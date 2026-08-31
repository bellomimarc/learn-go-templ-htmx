package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/marcello/saas-poc/internal/features/dashboard"
	"github.com/marcello/saas-poc/internal/features/system"
	"github.com/marcello/saas-poc/internal/features/website"
	"github.com/marcello/saas-poc/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
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

	website.RegisterRoutes(router)
	dashboard.RegisterRoutes(router, dashboard.NewTodoStore(pool))
	system.RegisterRoutes(router)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println(`
╔════════════════════════════════════════════════════════╗
║       SaaS Gestionale PoC - Minimal Stack 2026        ║
╚════════════════════════════════════════════════════════╝

📡 Server running on http://localhost:8080

📊 Routes:
	GET  /                → Website landing (Full HTML page with Templ)
	GET  /dashboard       → Dashboard (Full HTML page with Templ)
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
	• TODO handlers use an injected pgx connection pool and return Templ fragments
   • Classic REST calls to /api/info → server returns JSON with stdlib
	• Website landing and dashboard use separate Templ layouts

Press Ctrl+C to stop the server.
	`)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
