package system

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type StatusResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

type InfoResponse struct {
	Application string            `json:"application"`
	Version     string            `json:"version"`
	Stack       map[string]string `json:"stack"`
	Environment string            `json:"environment"`
}

var startTime = time.Now()

func RegisterRoutes(router chi.Router) {
	router.Get("/api/info", handleInfo)
	router.Get("/health", handleHealthCheck)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := InfoResponse{
		Application: "SaaS PoC",
		Version:     "0.1.0-alpha",
		Stack: map[string]string{
			"router":   "github.com/go-chi/chi/v5",
			"template": "github.com/a-h/templ",
			"frontend": "HTMX (unpkg CDN)",
			"encoding": "encoding/json (stdlib)",
		},
		Environment: "development",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v\n", err)
	}
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	uptime := time.Since(startTime).String()
	response := StatusResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    uptime,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v\n", err)
	}
}
