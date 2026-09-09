package dashboard

import (
	"log"
	"net/http"
	"time"

	dashboardviews "github.com/marcello/saas-poc/internal/delivery/dashboard/views"
	"github.com/marcello/saas-poc/internal/todos"

	"github.com/go-chi/chi/v5"
)

// statusLimiter enforces rate limiting on the /dashboard/status endpoint
// Limit: 5 requests per 2 seconds per IP
var statusLimiter = NewRateLimiter(5, 2*time.Second)
var eventBroker = newSSEBroker()

func RegisterRoutes(router chi.Router, todoService todos.TodoService, authenticate func(http.Handler) http.Handler) {
	router.Group(func(protected chi.Router) {
		protected.Use(authenticate)
		protected.Get("/dashboard", handleDashboard)
		protected.Get("/dashboard/events", handleEventsPage)
		protected.Get("/dashboard/events/stream", handleEventsStream(eventBroker))
		protected.Post("/dashboard/events", handleEventTrigger(eventBroker))
		protected.Get("/dashboard/loan", handleLoanApplicationPage)
		protected.Get("/dashboard/todos", handleTodoPage(todoService))
		protected.Get("/dashboard/status", handleStatus)
		protected.Post("/dashboard/loan/validate", handleLoanValidation)
		protected.Post("/dashboard/loan/submit", handleLoanSubmission)
		protected.Post("/dashboard/todos", handleTodoCreate(todoService))
		protected.Put("/dashboard/todos/{id}", handleTodoRename(todoService))
		protected.Patch("/dashboard/todos/{id}/toggle", handleTodoToggle(todoService))
		protected.Delete("/dashboard/todos/{id}", handleTodoDelete(todoService))
	})
}

// Shutdown ends long-lived dashboard streams so in-flight requests can drain.
func Shutdown() {
	eventBroker.shutdown()
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
