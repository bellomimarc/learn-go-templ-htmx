package middleware

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

const zitadelProjectRolesClaim = "urn:zitadel:iam:org:project:roles"

type OIDCAccessTokenVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func NewOIDCAccessTokenVerifier(ctx context.Context, issuer, clientID string) (*OIDCAccessTokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}

	return &OIDCAccessTokenVerifier{
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
	}, nil
}

func (verifier *OIDCAccessTokenVerifier) Verify(ctx context.Context, rawToken string) (Principal, error) {
	token, err := verifier.verifier.Verify(ctx, rawToken)
	if err != nil {
		return Principal{}, fmt.Errorf("verify access token: %w", err)
	}

	var claims struct {
		PreferredUsername string                     `json:"preferred_username"`
		Roles             map[string]json.RawMessage `json:"urn:zitadel:iam:org:project:roles"`
	}
	if err := token.Claims(&claims); err != nil {
		return Principal{}, fmt.Errorf("decode access token claims: %w", err)
	}

	roles := make([]string, 0, len(claims.Roles))
	for role := range claims.Roles {
		roles = append(roles, role)
	}

	return Principal{
		Subject:  token.Subject,
		Username: claims.PreferredUsername,
		Roles:    roles,
	}, nil
}
