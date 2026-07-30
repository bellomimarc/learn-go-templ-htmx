package dashboard

import (
	"log"
	"net/http"
	"time"

	dashboardviews "github.com/marcello/saas-poc/internal/features/dashboard/views"

	"github.com/go-chi/chi/v5"
)

// statusLimiter enforces rate limiting on the /api/status endpoint
// Limit: 5 requests per 2 seconds per IP
var statusLimiter = NewRateLimiter(5, 2*time.Second)

func RegisterRoutes(router chi.Router) {
	router.Get("/dashboard", handleDashboard)
	router.Get("/dashboard/loan", handleLoanApplicationPage)
	router.Get("/api/status", handleStatus)
	router.Post("/api/loan/validate", handleLoanValidation)
	router.Post("/api/loan/submit", handleLoanSubmission)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))
	if err := dashboardviews.Dashboard(locale).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering dashboard", http.StatusInternalServerError)
		log.Printf("Error rendering dashboard: %v\n", err)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))

	// Get client IP for rate limiting
	clientIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		clientIP = xff
	}

	// Check rate limit
	if !statusLimiter.IsAllowed(clientIP) {
		w.WriteHeader(http.StatusTooManyRequests)
		if err := dashboardviews.RateLimitComponent(locale).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering rate limit component", http.StatusInternalServerError)
			log.Printf("Error rendering rate limit component: %v\n", err)
		}
		return
	}

	status := locale.Text("status.online")
	timestamp := time.Now()
	time.Sleep(100 * time.Millisecond)

	if err := dashboardviews.StatusComponent(locale, status, timestamp).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering status component", http.StatusInternalServerError)
		log.Printf("Error rendering status component: %v\n", err)
	}
}
