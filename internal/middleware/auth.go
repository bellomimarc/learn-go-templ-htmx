package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type Principal struct {
	Subject  string
	Username string
	Roles    []string
}

type AccessTokenVerifier interface {
	Verify(context.Context, string) (Principal, error)
}

type principalContextKey struct{}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

func RequireBearer(verifier AccessTokenVerifier, allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r.Header.Get("Authorization"))
			if err != nil {
				writeAuthenticationRequired(w, r)
				return
			}

			principal, err := verifier.Verify(r.Context(), token)
			if err != nil {
				writeAuthenticationRequired(w, r)
				return
			}
			if len(allowed) > 0 && !hasAllowedRole(principal.Roles, allowed) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errors.New("missing or malformed bearer token")
	}
	return parts[1], nil
}

func hasAllowedRole(roles []string, allowed map[string]struct{}) bool {
	for _, role := range roles {
		if _, ok := allowed[role]; ok {
			return true
		}
	}
	return false
}

func writeAuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("WWW-Authenticate", `Bearer realm="dashboard"`)
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(authenticationBridge))
}

const authenticationBridge = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Signing in</title></head>
<body>
<script>
(() => {
  const token = sessionStorage.getItem("zitadel_access_token");
  const target = window.location.pathname + window.location.search;
  const login = () => window.location.replace("/auth/login?return_to=" + encodeURIComponent(target));
  if (!token) {
    login();
    return;
  }
  fetch(target, {headers: {Authorization: "Bearer " + token, Accept: "text/html"}})
    .then(async response => {
      if (response.status === 401) {
        sessionStorage.removeItem("zitadel_access_token");
        sessionStorage.removeItem("zitadel_id_token");
        login();
        return;
      }
      const html = await response.text();
      document.open();
      document.write(html);
      document.close();
    })
    .catch(login);
})();
</script>
</body>
</html>`
