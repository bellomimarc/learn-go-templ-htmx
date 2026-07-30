package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/marcello/saas-poc/internal/features/dashboard"
	"github.com/marcello/saas-poc/internal/features/system"
	"github.com/marcello/saas-poc/internal/features/website"
	"github.com/marcello/saas-poc/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	router := chi.NewRouter()

	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(middleware.Logging)

	website.RegisterRoutes(router)
	dashboard.RegisterRoutes(router)
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
   GET  /api/status      → Status component (HTMX endpoint, returns HTML fragment)
	POST /api/loan/validate → Loan form cross-field validation (HTML fragment)
	POST /api/loan/submit  → Loan form submission simulation (HTML fragment)
   GET  /api/info        → API Info (JSON REST endpoint)
   GET  /health          → Health check (JSON)

🛠️  Tech Stack:
   • Router: github.com/go-chi/chi/v5
   • Templates: github.com/a-h/templ (type-safe HTML)
   • Frontend: HTMX (CDN) + Bootstrap styling
   • REST API: encoding/json (stdlib)
   • Middleware: Chi + custom handlers

🔐 Security & Compliance:
   • No secrets hardcoded
   • Minimal external dependencies (2 deps only)
   • Type-safe HTML rendering (Templ prevents XSS)
   • Standard library for JSON encoding

💡 Architecture:
   • HTMX talks to /api/status → server renders partial HTML → browser updates DOM
   • Classic REST calls to /api/info → server returns JSON with stdlib
	• Website landing and dashboard use separate Templ layouts

Press Ctrl+C to stop the server.
	`)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
