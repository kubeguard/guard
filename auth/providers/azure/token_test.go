/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
