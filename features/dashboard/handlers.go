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
	router.Get("/api/status", handleStatus)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dashboardviews.Dashboard().Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering dashboard", http.StatusInternalServerError)
		log.Printf("Error rendering dashboard: %v\n", err)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	status := "Online"
	timestamp := time.Now()
	time.Sleep(500 * time.Millisecond)

	if err := dashboardviews.StatusComponent(status, timestamp).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering status component", http.StatusInternalServerError)
		log.Printf("Error rendering status component: %v\n", err)
	}
}
