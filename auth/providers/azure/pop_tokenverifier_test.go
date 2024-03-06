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

package azure_test

import (
	"fmt"
	"testing"
	"time"

	"go.kubeguard.dev/guard/auth/providers/azure"

	"github.com/stretchr/testify/assert"
)

func TestPopTokenVerifier_Verify(t *testing.T) {
	verifier, err := azure.NewPoPVerifier("testHostname", 15*time.Minute, 1*time.Minute)
	assert.NoError(t, err)

	// Test cases where no error is expected
	noErrorTestCases := []struct {
		desc string
		kid  string
	}{
		{
			desc: "happy path test case, all arguments are passed correctly",
		},
		{
			desc: "'ts' claim is passed as string",
			kid:  azure.TsClaimsTypeString,
		},
	}
	for _, tC := range noErrorTestCases {
		t.Run(tC.desc, func(t *testing.T) {
			validToken, _ := azure.NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(tC.kid).GetToken()
			_, err := verifier.ValidatePopToken(validToken)
			assert.NoError(t, err)
		})
	}

	// Test cases asserting "ErrorEquals"
	testCases := []struct {
		desc      string
		kid       string
		hostname  string
		errString string
	}{
		{
			desc:      "PoP token is not in the right format",
			kid:       azure.BadTokenKey,
			errString: "PoP token invalid schema. Token length: 1",
		},
		{
			desc:      "PoP token is in the right format but is incorrect and could not be parsed",
			kid:       fmt.Sprintf("%s.%s.%s", azure.BadTokenKey, azure.BadTokenKey, azure.BadTokenKey),
			hostname:  "testHostname",
			errString: "Could not parse PoP token. Error: invalid character 'm' looking for beginning of value",
		},
		{
			desc:      "'keyID' claim in the header is missing",
			kid:       azure.HeaderBadKeyID,
			errString: "No KeyID found in PoP token header",
		},
		{
			desc:      "'algo' claim in the header is incorrect",
			kid:       azure.HeaderBadAlgo,
			errString: "Wrong algorithm found in PoP token header, expected 'RS256' having 'wrong'",
		},
		{
			desc:      "'typ' claim in the header is incorrect",
			kid:       azure.HeaderBadtyp,
			errString: "Wrong typ. Expected 'pop' having 'wrong'",
		},
		{
			desc:      "'typ' claim is not present in the header",
			kid:       azure.HeaderBadtypMissing,
			errString: "Invalid token. 'typ' claim is missing",
		},
		{
			desc:      "Invalid token. 'typ' claim should be of string",
			kid:       azure.HeaderBadtypType,
			hostname:  "testHostname",
			errString: "Invalid token. 'typ' claim should be of string",
		},
		{
			desc:      "'ts' claim is not present in the payload",
			kid:       azure.TsClaimsMissing,
			errString: "Invalid token. 'ts' claim is missing",
		},
		{
			desc:      "Request and validation for the PoP token are running for different hostnames",
			hostname:  "wrongHostnme",
			errString: "Invalid Pop token due to host mismatch. Expected: \"testHostname\", received: \"wrongHostnme\"",
		},
		{
			desc:      "'cnf' is not present in the access token",
			kid:       azure.AtCnfClaimMissing,
			hostname:  "testHostname",
			errString: "could not retrieve 'cnf' claim from access token",
		},
		{
			desc:      "'cnf' in the access token does not have the right value",
			kid:       azure.AtCnfClaimWrong,
			hostname:  "testHostname",
			errString: "PoP token validate failed: 'cnf' claim mismatch",
		},
		{
			desc:      "'cnf' claim in the payload is not present",
			kid:       azure.CnfClaimsMissing,
			hostname:  "testHostname",
			errString: "Invalid token. 'cnf' claim is missing",
		},
		{
			desc:      "'cnf' claim in the payload does not have 'jwk'",
			kid:       azure.CnfJwkClaimsMissing,
			hostname:  "testHostname",
			errString: "Invalid token. 'cnf' claim is not in expected format",
		},
		{
			desc:      "'jwk' in 'cnf' claim in the payload is empty",
			kid:       azure.CnfJwkClaimsEmpty,
			hostname:  "testHostname",
			errString: "Invalid token. 'jwk' claim is empty",
		},
		{
			desc:      "'u' claim is not present in the payload",
			kid:       azure.UClaimsMissing,
			hostname:  "testHostname",
			errString: "Invalid token. 'u' claim is missing",
		},
		{
			desc:      "'u' claim in the payload is not of string type",
			kid:       azure.UClaimsWrongType,
			hostname:  "testHostname",
			errString: "Invalid token. 'u' claim should be of string",
		},
		{
			desc:      "'at' claim in the payload is not of type string",
			kid:       azure.AtClaimsWrongType,
			hostname:  "testHostname",
			errString: "Invalid token. 'at' claim should be string",
		},
		{
			desc:      "'at' claim is not present in the payload",
			kid:       azure.AtClaimsMissing,
			hostname:  "testHostname",
			errString: "Invalid token. access token missing",
		},
		{
			desc:      "'nonce' claim in the payload is missing",
			kid:       azure.NonceClaimMissing,
			hostname:  "testHostname",
			errString: "Invalid token. 'nonce' claim is missing",
		},
		{
			desc:      "'nonce' claim in the payload is not a string",
			kid:       azure.NonceClaimNotString,
			hostname:  "testHostname",
			errString: "Invalid token. 'nonce' claim should be of type string",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			invalidToken, _ := azure.NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName(tC.hostname).SetKid(tC.kid).GetToken()
			_, err := verifier.ValidatePopToken(invalidToken)
			assert.EqualError(t, err, tC.errString)
		})
	}

	// Test cases asserting ErrorContains
	testCasesContainsErrors := []struct {
		desc      string
		kid       string
		ts        int64
		hostname  string
		errString string
	}{
		{
			desc:      "'jwk' in the 'cnf' claim in the payload is not of string type",
			kid:       azure.CnfJwkClaimsWrong,
			errString: "failed while parsing 'jwk' claim in PoP token",
		},
		{
			desc:      "'jwk' in the 'cnf' claim in the access token is not of string type",
			kid:       azure.AccessTokenCnfWrong,
			hostname:  "testHostname",
			errString: "failed while parsing 'cnf' in access token",
		},
		{
			desc:      "'at' claim value in they payload is not correct and could not be parsed",
			kid:       azure.AtClaimIncorrect,
			hostname:  "testHostname",
			errString: "could not parse access token in PoP token",
		},
		{
			desc:      "RSA verify error due to invalid signature in the PoP token",
			kid:       azure.SignatureWrongType,
			hostname:  "testHostname",
			errString: "RSA verify err: crypto/rsa: verification error",
		},
	}
	for _, tC := range testCasesContainsErrors {
		t.Run(tC.desc, func(t *testing.T) {
			invalidToken, _ := azure.NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(tC.kid).GetToken()
			_, err := verifier.ValidatePopToken(invalidToken)
			assert.ErrorContains(t, err, tC.errString)
		})
	}

	t.Run("'ts' has an expired timestamp", func(t *testing.T) {
		expiredToken, _ := azure.NewPoPTokenBuilder().SetTimestamp(time.Now().Add(time.Minute * -20).Unix()).GetToken()
		_, err := verifier.ValidatePopToken(expiredToken)
		assert.NotNilf(t, err, "PoP verification succeed.")
		assert.Containsf(t, err.Error(), "Token is expired", "Error message is not as expected")
	})

	t.Run("'ts' claim in the payload is of unknown type and cannot be parsed. 'ts' is set to the default timestamp which has expired", func(t *testing.T) {
		invalidToken, _ := azure.NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(azure.TsClaimsTypeUnknown).GetToken()
		_, err := verifier.ValidatePopToken(invalidToken)
		assert.Containsf(t, err.Error(), "Token is expired", "Error message is not as expected")
	})

	t.Run("'nonce' claim been reused - Cache validation", func(t *testing.T) {
		// Setting up the verifier expiration time to 4 seconds and cache expiration time to -3 seconds.
		// This will ensure that the cache is expired before the verifier on the second call after 1sec.
		nonceVerifier, err := azure.NewPoPVerifier("testHostname", 4*time.Second, 1*time.Second)
		assert.NoError(t, err)
		// Generating a first valid token with a nonce claim hard coded to be reused. This should pass the validation.
		validToken, _ := azure.NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(azure.NonceClaimHardcoded).GetToken()
		_, err = nonceVerifier.ValidatePopToken(validToken)
		assert.NoError(t, err)
		// Generating a second valid token with a nonce claim hard coded to be reused. This should fail the validation because the value is cached.
		validToken, _ = azure.NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(azure.NonceClaimHardcoded).GetToken()
		_, err = nonceVerifier.ValidatePopToken(validToken)
		assert.Containsf(t, err.Error(), "Invalid token. 'nonce' claim is reused", "Error message is not as expected")
		// Sleeping for 3 seconds to ensure that the cache is expired before the verifier is called again.
		time.Sleep(3 * time.Second)
		_, err = nonceVerifier.ValidatePopToken(validToken)
		assert.NoError(t, err)
	})
}
