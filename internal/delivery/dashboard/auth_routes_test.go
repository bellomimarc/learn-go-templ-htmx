package dashboard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutesProtectsEveryDashboardRoute(t *testing.T) {
	router := chi.NewRouter()
	authenticationCalls := 0
	RegisterRoutes(router, nil, func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			authenticationCalls++
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	})

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/dashboard"},
		{http.MethodGet, "/dashboard/events"},
		{http.MethodGet, "/dashboard/events/stream"},
		{http.MethodPost, "/dashboard/events"},
		{http.MethodGet, "/dashboard/loan"},
		{http.MethodPost, "/dashboard/loan/validate"},
		{http.MethodPost, "/dashboard/loan/submit"},
		{http.MethodGet, "/dashboard/todos"},
		{http.MethodPost, "/dashboard/todos"},
		{http.MethodPut, "/dashboard/todos/1"},
		{http.MethodPatch, "/dashboard/todos/1/toggle"},
		{http.MethodDelete, "/dashboard/todos/1"},
		{http.MethodGet, "/dashboard/status"},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
			}
		})
	}

	if authenticationCalls != len(tests) {
		t.Fatalf("expected authentication on %d routes, got %d", len(tests), authenticationCalls)
	}
}
