package dashboard

import (
	"log"
	"net/http"
	"time"

	dashboardviews "github.com/marcello/saas-poc/features/dashboard/views"

	"github.com/go-chi/chi/v5"
)

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

	status := locale.Text("status.online")
	timestamp := time.Now()
	time.Sleep(500 * time.Millisecond)

	if err := dashboardviews.StatusComponent(locale, status, timestamp).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering status component", http.StatusInternalServerError)
		log.Printf("Error rendering status component: %v\n", err)
	}
}
