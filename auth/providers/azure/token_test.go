package azure

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestExtractTokenClaimsWithValidToken(t *testing.T) {
	mySigningKey := []byte("AllYourBase")

	// Create the Claims
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Unix(1516239022, 0)),
		Issuer:    "test",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		t.Fatal(err)
	}

	extractedClaims, err := extractTokenClaims(ss)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("claims: %s", extractedClaims)
}

func TestExtractTokenClaimsWithInvalidToken(t *testing.T) {
	_, err := extractTokenClaims("")
	if err == nil {
		t.Fatal("expected to see the error with invalid token")
	}
}
