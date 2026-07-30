package website

import (
	"log"
	"net/http"

	websiteviews "github.com/marcello/saas-poc/internal/features/website/views"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router) {
	router.Get("/", handleLanding)
}

func handleLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := websiteviews.Landing().Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering website", http.StatusInternalServerError)
		log.Printf("Error rendering website: %v\n", err)
	}
}
