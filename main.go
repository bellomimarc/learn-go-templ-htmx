package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/marcello/saas-poc/features/dashboard"
	"github.com/marcello/saas-poc/features/system"
	"github.com/marcello/saas-poc/features/website"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ===============================================
// Middleware (Custom Logic Layer)
// ===============================================

// loggingMiddleware è un middleware custom che logga informazioni della richiesta.
// Chi fornisce già middleware standard (Logger, Recoverer), ma questo dimostra
// come estendere con custom logic.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// ===============================================
// Main: Setup Router e Start Server
// ===============================================

func main() {
	// Creiamo un router Chi.
	// Chi è un router HTTP minimalista che supporta middleware e nested routes.
	router := chi.NewRouter()

	// Middleware globali (applicati a tutte le rotte).
	// Logger e Recoverer sono middleware standard forniti da Chi.
	router.Use(middleware.Logger)    // Logga ogni richiesta HTTP
	router.Use(middleware.Recoverer) // Recupera da panic durante request handling
	router.Use(loggingMiddleware)    // Custom middleware

	// Feature-level route registration keeps handlers and views grouped by domain.
	website.RegisterRoutes(router)
	dashboard.RegisterRoutes(router)
	system.RegisterRoutes(router)

	// ============ Server Start ============

	// Configuriamo il server HTTP.
	// Usiamo la stdlib net/http per minimizzare dipendenze esterne.
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
   GET  /api/status      → Status component (HTMX endpoint, returns HTML fragment)
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

	// Avviamo il server.
	// ListenAndServe è bloccante fino a errore o shutdown.
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
