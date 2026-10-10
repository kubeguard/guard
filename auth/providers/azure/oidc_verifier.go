package azure

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc"
)

var (
	_ AccessTokenVerifier = (*OIDCAccessTokenVerifier)(nil)
	_ VerifiedAccessToken = (*oidcVerifiedAccessToken)(nil)
)

type OIDCAccessTokenVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func (o *OIDCAccessTokenVerifier) Verify(ctx context.Context, rawAccessToken string) (VerifiedAccessToken, error) {
	token, err := o.verifier.Verify(ctx, rawAccessToken)
	if err != nil {
		return nil, err
	}

	return &oidcVerifiedAccessToken{token: token}, nil
}

type oidcVerifiedAccessToken struct {
	token *oidc.IDToken
}

func (t *oidcVerifiedAccessToken) Claims() (claims, error) {
	if t.token == nil {
		return nil, fmt.Errorf("claims not set")
	}

	parsedClaims := claims{}
	if err := t.token.Claims(&parsedClaims); err != nil {
		return nil, err
	}

	return parsedClaims, nil
}
