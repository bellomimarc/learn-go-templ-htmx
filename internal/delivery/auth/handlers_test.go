package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestNewConfigBuildsZitadelEndpoints(t *testing.T) {
	config, err := NewConfig("http://auth.localhost:8081/", "client-id", "http://localhost:8080/")
	if err != nil {
		t.Fatalf("create auth config: %v", err)
	}

	if config.AuthorizationEndpoint != "http://auth.localhost:8081/oauth/v2/authorize" {
		t.Fatalf("unexpected authorization endpoint: %q", config.AuthorizationEndpoint)
	}
	if config.RedirectURI != "http://localhost:8080/auth/callback" {
		t.Fatalf("unexpected redirect URI: %q", config.RedirectURI)
	}
	if config.PostLogoutRedirectURI != "http://localhost:8080/" {
		t.Fatalf("unexpected post-logout redirect URI: %q", config.PostLogoutRedirectURI)
	}
}

func TestNewConfigRejectsMissingClientID(t *testing.T) {
	_, err := NewConfig("http://auth.localhost:8081", "", "http://localhost:8080")
	if err == nil {
		t.Fatal("expected missing client ID to fail")
	}
}

func TestAuthConfigEndpoint(t *testing.T) {
	config, err := NewConfig("http://auth.localhost:8081", "client-id", "http://localhost:8080")
	if err != nil {
		t.Fatalf("create auth config: %v", err)
	}
	router := chi.NewRouter()
	RegisterRoutes(router, config)

	request := httptest.NewRequest(http.MethodGet, "/auth/config", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store response, got %q", recorder.Header().Get("Cache-Control"))
	}
	var response Config
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode config response: %v", err)
	}
	if response.ClientID != "client-id" || response.TokenEndpoint != "http://auth.localhost:8081/oauth/v2/token" {
		t.Fatalf("unexpected config response: %+v", response)
	}
}
