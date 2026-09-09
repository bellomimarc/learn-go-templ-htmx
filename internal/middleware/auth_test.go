package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubAccessTokenVerifier struct {
	principal Principal
	err       error
	token     string
}

func (verifier *stubAccessTokenVerifier) Verify(_ context.Context, token string) (Principal, error) {
	verifier.token = token
	return verifier.principal, verifier.err
}

func TestRequireBearerRejectsMissingToken(t *testing.T) {
	handler := RequireBearer(&stubAccessTokenVerifier{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler should not run")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/dashboard", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	if recorder.Header().Get("WWW-Authenticate") != `Bearer realm="dashboard"` {
		t.Fatalf("unexpected authentication challenge: %q", recorder.Header().Get("WWW-Authenticate"))
	}
}

func TestRequireBearerReturnsBrowserAuthenticationBridge(t *testing.T) {
	handler := RequireBearer(&stubAccessTokenVerifier{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	request := httptest.NewRequest(http.MethodGet, "/dashboard?lang=it", nil)
	request.Header.Set("Accept", "text/html")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	for _, expected := range []string{"sessionStorage", "/auth/login", "window.location.search"} {
		if !strings.Contains(recorder.Body.String(), expected) {
			t.Fatalf("expected browser bridge to contain %q", expected)
		}
	}
}

func TestRequireBearerRejectsInvalidToken(t *testing.T) {
	verifier := &stubAccessTokenVerifier{err: errors.New("invalid token")}
	handler := RequireBearer(verifier)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler should not run")
	}))
	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Authorization", "Bearer invalid")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	if verifier.token != "invalid" {
		t.Fatalf("expected verifier to receive token, got %q", verifier.token)
	}
}

func TestRequireBearerRejectsPrincipalWithoutAllowedRole(t *testing.T) {
	verifier := &stubAccessTokenVerifier{principal: Principal{Subject: "user-1", Roles: []string{"viewer"}}}
	handler := RequireBearer(verifier, "super-admin", "regular-user")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler should not run")
	}))
	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Authorization", "Bearer valid")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestRequireBearerAddsAuthorizedPrincipalToContext(t *testing.T) {
	expected := Principal{Subject: "user-1", Username: "regular-user@example.test", Roles: []string{"regular-user"}}
	verifier := &stubAccessTokenVerifier{principal: expected}
	handler := RequireBearer(verifier, "super-admin", "regular-user")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actual, ok := PrincipalFromContext(r.Context())
		if !ok || actual.Subject != expected.Subject || actual.Username != expected.Username {
			t.Fatalf("unexpected principal in context: %+v, present=%t", actual, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Authorization", "bearer valid")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}
