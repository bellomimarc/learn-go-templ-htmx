package middleware

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestZitadelProjectRolesClaimShape(t *testing.T) {
	rawClaims := []byte(`{
		"preferred_username": "super-admin@example.test",
		"urn:zitadel:iam:org:project:roles": {
			"super-admin": {"organization-id": "zitadel.localhost"},
			"auditor": {"organization-id": "zitadel.localhost"}
		}
	}`)

	var claims struct {
		PreferredUsername string                     `json:"preferred_username"`
		Roles             map[string]json.RawMessage `json:"urn:zitadel:iam:org:project:roles"`
	}
	if err := json.Unmarshal(rawClaims, &claims); err != nil {
		t.Fatalf("decode Zitadel claims: %v", err)
	}

	roles := make([]string, 0, len(claims.Roles))
	for role := range claims.Roles {
		roles = append(roles, role)
	}

	if claims.PreferredUsername != "super-admin@example.test" {
		t.Fatalf("unexpected preferred username: %q", claims.PreferredUsername)
	}
	if !slices.Contains(roles, "super-admin") || !slices.Contains(roles, "auditor") {
		t.Fatalf("unexpected roles: %v", roles)
	}
}
