package auth

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	authviews "github.com/marcello/saas-poc/internal/delivery/auth/views"

	"github.com/go-chi/chi/v5"
)

//go:embed static/client.js
var clientJavaScript []byte

type Config struct {
	Issuer                string   `json:"issuer"`
	ClientID              string   `json:"clientId"`
	AuthorizationEndpoint string   `json:"authorizationEndpoint"`
	TokenEndpoint         string   `json:"tokenEndpoint"`
	EndSessionEndpoint    string   `json:"endSessionEndpoint"`
	RedirectURI           string   `json:"redirectUri"`
	PostLogoutRedirectURI string   `json:"postLogoutRedirectUri"`
	Scopes                []string `json:"scopes"`
}

func NewConfig(issuer, clientID, applicationOrigin string) (Config, error) {
	issuer = strings.TrimRight(issuer, "/")
	applicationOrigin = strings.TrimRight(applicationOrigin, "/")
	if err := validateHTTPURL("ZITADEL_ISSUER", issuer); err != nil {
		return Config{}, err
	}
	if clientID == "" {
		return Config{}, fmt.Errorf("ZITADEL_CLIENT_ID is required")
	}
	if err := validateHTTPURL("APPLICATION_ORIGIN", applicationOrigin); err != nil {
		return Config{}, err
	}

	return Config{
		Issuer:                issuer,
		ClientID:              clientID,
		AuthorizationEndpoint: issuer + "/oauth/v2/authorize",
		TokenEndpoint:         issuer + "/oauth/v2/token",
		EndSessionEndpoint:    issuer + "/oidc/v1/end_session",
		RedirectURI:           applicationOrigin + "/auth/callback",
		PostLogoutRedirectURI: applicationOrigin + "/",
		Scopes:                []string{"openid", "profile", "email"},
	}, nil
}

func RegisterRoutes(router chi.Router, config Config) {
	router.Get("/auth/config", handleConfig(config))
	router.Get("/auth/client.js", handleClientJavaScript)
	router.Get("/auth/login", handlePage("login"))
	router.Get("/auth/callback", handlePage("callback"))
	router.Get("/auth/logout", handlePage("logout"))
}

func handleConfig(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(config); err != nil {
			log.Printf("write authentication config: %v", err)
		}
	}
}

func handleClientJavaScript(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	_, _ = w.Write(clientJavaScript)
}

func handlePage(flow string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'unsafe-inline'; connect-src 'self' http://auth.localhost:8081; base-uri 'none'; frame-ancestors 'none'")
		if err := authviews.Page(flow).Render(r.Context(), w); err != nil {
			http.Error(w, "Unable to render authentication page", http.StatusInternalServerError)
			log.Printf("render authentication page: %v", err)
		}
	}
}

func validateHTTPURL(name, value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%s must be an absolute HTTP(S) URL without query or fragment", name)
	}
	return nil
}
