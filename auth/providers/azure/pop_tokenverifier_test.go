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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
)

const (
	popAccessToken           = `{ "aud": "client", "iss" : "kd", "exp" : "%d","cnf": {"kid":"%s","xms_ksl":"sw"} }`
	popAccessTokenWithoutCnf = `{ "aud": "client", "iss" : "kd", "exp" : "%d" }`
)

type swPoPKey struct {
	key    *rsa.PrivateKey
	keyID  string
	jwk    string
	jwkTP  string
	reqCnf string
}

func (swk *swPoPKey) Alg() string {
	return "RS256"
}

func (swk *swPoPKey) Sign(payload []byte) ([]byte, error) {
	return swk.key.Sign(rand.Reader, payload, crypto.SHA256)
}

func (swk *swPoPKey) KeyID() string {
	return swk.keyID
}

func (swk *swPoPKey) Cnf() string {
	return swk.reqCnf
}

func (swk *swPoPKey) Jwk() string {
	return swk.jwk
}

func NewSWPoPKey() (*swPoPKey, error) {
	pop := &swPoPKey{}
	rsa, err := rsa.GenerateKey(rand.Reader, 1028)
	if err != nil {
		return nil, err
	}
	pop.key = rsa
	pubKey := rsa.PublicKey
	e := big.NewInt(int64(pubKey.E))
	eB64 := base64.RawURLEncoding.EncodeToString(e.Bytes())
	n := pubKey.N
	nB64 := base64.RawURLEncoding.EncodeToString(n.Bytes())
	jwk := fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`, eB64, nB64)
	jwkS256 := sha256.Sum256([]byte(jwk))
	pop.jwkTP = base64.RawURLEncoding.EncodeToString(jwkS256[:])

	reqCnfJSON := fmt.Sprintf(`{"kid":"%s","xms_ksl":"sw"}`, pop.jwkTP)
	pop.reqCnf = base64.RawURLEncoding.EncodeToString([]byte(reqCnfJSON))
	pop.keyID = pop.jwkTP
	pop.jwk = fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s","alg":"RS256","kid":"%s"}`, eB64, nB64, pop.keyID)

	return pop, nil
}

type swKey struct {
	keyID  string
	pKey   *rsa.PrivateKey
	pubKey interface{}
}

func (swk *swKey) Alg() string {
	return "RS256"
}

func (swk *swKey) KeyID() string {
	return ""
}

func NewSwkKey() (*swKey, error) {
	rsa, err := rsa.GenerateKey(rand.Reader, 1028)
	if err != nil {
		return nil, err
	}
	return &swKey{"", rsa, rsa.Public()}, nil
}

func (swk *swKey) GenerateToken(payload []byte) (string, error) {
	pKey := &jose.JSONWebKey{Key: swk.pKey, Algorithm: swk.Alg(), KeyID: swk.KeyID()}

	// create a Square.jose RSA signer, used to sign the JWT
	signerOpts := jose.SignerOptions{}
	signerOpts.WithType("JWT")
	rsaSigner, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: pKey}, &signerOpts)
	if err != nil {
		return "", err
	}
	jws, err := rsaSigner.Sign(payload)
	if err != nil {
		return "", err
	}

	token, err := jws.CompactSerialize()
	if err != nil {
		return "", err
	}
	return token, nil
}

const (
	badTokenKey         = "badToken"
	headerBadKeyID      = "headerBadKeyID"
	headerBadAlgo       = "headerBadAlgo"
	headerBadtyp        = "headerBadtyp"
	headerBadtypType    = "headerBadtypType"
	headerBadtypMissing = "headerBadtypMissing"
	uClaimsMissing      = "uClaimsMissing"
	tsClaimsMissing     = "tsClaimsMissing"
	atClaimsMissing     = "atClaimsMissing"
	atClaimIncorrect    = "atClaimIncorrect"
	cnfClaimsMissing    = "cnfClaimsMissing"
	cnfJwkClaimsEmpty   = "cnfJwkClaimsEmpty"
	cnfJwkClaimsWrong   = "cnfJwkClaimsWrong"
	cnfJwkClaimsMissing = "cnfJwkClaimsMissing"
	accessTokenCnfWrong = "accessTokenCnfWrong"
	atClaimsWrongType   = "atClaimsWrongType"
	atCnfClaimMissing   = "atCnfClaimMissing"
	atCnfClaimWrong     = "atCnfClaimWrong"
	tsClaimsTypeString  = "tsClaimsTypeString"
	tsClaimsTypeUnknown = "tsClaimsTypeUnknown"
	uClaimsWrongType    = "uClaimsWrongType"
	signatureWrongType  = "signatureWrongType"
)

// A struct that represents a PoP token
type PoPToken struct {
	Header    string
	Payload   string
	Signature string
}

// A concrete builder struct that implements the steps to build a PoP token
type PoPTokenBuilderImpl struct {
	popKey   *swPoPKey
	swkKey   *swKey
	ts       int64
	hostName string
	kid      string // used for testing purposes
	token    PoPToken
}

// A constructor function that returns a new PoPTokenBuilderImpl
func NewPoPTokenBuilder() *PoPTokenBuilderImpl {
	return &PoPTokenBuilderImpl{}
}

func (b *PoPTokenBuilderImpl) SetTimestamp(ts int64) *PoPTokenBuilderImpl {
	b.ts = ts
	return b
}

func (b *PoPTokenBuilderImpl) SetKid(kid string) *PoPTokenBuilderImpl {
	b.kid = kid
	return b
}

func (b *PoPTokenBuilderImpl) SetHostName(hostName string) *PoPTokenBuilderImpl {
	b.hostName = hostName
	return b
}

// A method that sets the header of the PoP token
func (b *PoPTokenBuilderImpl) SetHeader() error {
	keyID := b.popKey.KeyID()
	algo := b.popKey.Alg()
	typ := "pop"
	if b.kid == headerBadKeyID {
		keyID = ""
	}
	if b.kid == headerBadAlgo {
		algo = "wrong"
	}
	if b.kid == headerBadtyp {
		typ = "wrong"
	}
	header := fmt.Sprintf(`{"typ":"%s","alg":"%s","kid":"%s"}`, typ, algo, keyID)

	if b.kid == headerBadtypType {
		wrongTyp := 1
		header = fmt.Sprintf(`{"typ":%d,"alg":"%s","kid":"%s"}`, wrongTyp, algo, keyID)
	}
	if b.kid == headerBadtypMissing {
		header = fmt.Sprintf(`{"alg":"%s","kid":"%s"}`, algo, keyID)
	}
	b.token.Header = base64.RawURLEncoding.EncodeToString([]byte(header))
	return nil
}

// A method that sets the payload of the PoP token
func (b *PoPTokenBuilderImpl) SetPayload() error {
	var cnf string
	if b.kid == atCnfClaimWrong {
		cnf = "wrongCnf"
	} else {
		cnf = b.popKey.KeyID()
	}

	nonce := uuid.New().String()
	nonce = strings.Replace(nonce, "-", "", -1)

	accessTokenData := fmt.Sprintf(popAccessToken, time.Now().Add(time.Minute*5).Unix(), cnf)
	if b.kid == atCnfClaimMissing {
		accessTokenData = fmt.Sprintf(popAccessTokenWithoutCnf, time.Now().Add(time.Minute*5).Unix())
	}
	if b.kid == accessTokenCnfWrong {
		accessTokenData = fmt.Sprintf(`{ "aud": "client", "iss" : "kd", "exp" : "%d","cnf": {"kid":1} }`, time.Now().Add(time.Minute*5).Unix())
	}

	at, err := b.swkKey.GenerateToken([]byte(accessTokenData))
	if err != nil {
		return fmt.Errorf("Error when generating token. Error:%+v", err)
	}

	payload := fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, b.ts, b.hostName, b.popKey.Jwk(), nonce)
	if b.kid == tsClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "u": "%d", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, 1, b.popKey.Jwk(), nonce)
	}
	if b.kid == uClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "cnf":{"jwk":%s}, "nonce":"%s"}`, at, b.ts, b.popKey.Jwk(), nonce)
	}
	if b.kid == cnfJwkClaimsEmpty {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{}, "nonce":"%s"}`, at, b.ts, b.hostName, nonce)
	}
	if b.kid == cnfJwkClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":1, "nonce":"%s"}`, at, b.ts, b.hostName, nonce)
	}
	if b.kid == cnfJwkClaimsWrong {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":1}, "nonce":"%s"}`, at, b.ts, b.hostName, nonce)
	}
	if b.kid == cnfClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "nonce": "%s"}`, at, b.ts, b.hostName, nonce)
	}
	if b.kid == tsClaimsTypeString {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : "%s", "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, strconv.FormatInt(b.ts, 10), b.hostName, b.popKey.Jwk(), nonce)
	}
	if b.kid == tsClaimsTypeUnknown {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %t, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, bool(true), b.hostName, b.popKey.Jwk(), nonce)
	}
	if b.kid == atClaimsWrongType {
		payload = fmt.Sprintf(`{ "at" : %d, "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, 12, b.ts, b.hostName, b.popKey.Jwk(), nonce)
	}
	if b.kid == uClaimsWrongType {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": %d, "cnf":{"jwk":%s}, "nonce":"%s"}`, at, b.ts, 1, b.popKey.Jwk(), nonce)
	}
	if b.kid == atClaimsMissing {
		payload = fmt.Sprintf(`{ "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, b.ts, b.hostName, b.popKey.Jwk(), nonce)
	}
	if b.kid == atClaimIncorrect {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, fmt.Sprintf("%s.%s.%s", badTokenKey, badTokenKey, badTokenKey), b.ts, b.hostName, b.popKey.Jwk(), nonce)
	}
	b.token.Payload = base64.RawURLEncoding.EncodeToString([]byte(payload))
	return nil
}

// A method that sets the signature of the PoP token
func (b *PoPTokenBuilderImpl) SetSignature() error {
	if b.kid == signatureWrongType {
		b.token.Signature = "wrongSignature"
		return nil
	}
	h256 := sha256.Sum256([]byte(b.token.Header + "." + b.token.Payload))
	signature, err := b.popKey.Sign(h256[:])
	if err != nil {
		return fmt.Errorf("Error while signing pop key. Error:%+v", err)
	}
	b.token.Signature = base64.RawURLEncoding.EncodeToString(signature)
	return nil
}

// A method that returns the final PoP token as a string
func (b *PoPTokenBuilderImpl) GetToken() (string, error) {
	var err error
	b.popKey, err = NewSWPoPKey()
	if err != nil {
		return "", fmt.Errorf("Failed to generate Pop key. Error:%+v", err)
	}
	b.swkKey, err = NewSwkKey()
	if err != nil {
		return "", fmt.Errorf("Failed to generate SF key. Error:%+v", err)
	}

	if strings.Contains(b.kid, badTokenKey) {
		return b.kid, nil
	}

	err = b.SetHeader()
	if err != nil {
		return "", err
	}
	err = b.SetPayload()
	if err != nil {
		return "", err
	}
	err = b.SetSignature()
	if err != nil {
		return "", err
	}

	finalToken := b.token.Header + "." + b.token.Payload + "." + b.token.Signature
	return finalToken, nil
}

func TestPopTokenVerifier_Verify(t *testing.T) {
	verifier := NewPoPVerifier("testHostname", 15*time.Minute)

	validToken, _ := NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").GetToken()
	_, err := verifier.ValidatePopToken(validToken)
	assert.NoError(t, err)

	// 'ts' claim is passed as string
	validToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(tsClaimsTypeString).GetToken()
	_, err = verifier.ValidatePopToken(validToken)
	assert.NoError(t, err)

	// PoP token is not in the right format
	invalidToken, _ := NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetKid(badTokenKey).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "PoP token invalid schema. Token length: 1")

	// PoP token is in the right format but is incorrect and could not be parsed
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(fmt.Sprintf("%s.%s.%s", badTokenKey, badTokenKey, badTokenKey)).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Could not parse PoP token. Error: invalid character 'm' looking for beginning of value")

	// 'keyID' claim in the header is missing
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetKid(headerBadKeyID).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "No KeyID found in PoP token header")

	// 'algo' claim in the header is incorrect
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetKid(headerBadAlgo).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Wrong algorithm found in PoP token header, expected 'RS256' having 'wrong'")

	// 'typ' claim in the header is incorrect
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetKid(headerBadtyp).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Wrong typ. Expected 'pop' having 'wrong'")

	// 'typ' claim is not present in the header
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetKid(headerBadtypMissing).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'typ' claim is missing")

	// 'typ' claim in the header is not of string type
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(headerBadtypType).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'typ' claim should be of string")

	// 'ts' has an expired timestamp
	expiredToken, _ := NewPoPTokenBuilder().SetTimestamp(time.Now().Add(time.Minute * -20).Unix()).GetToken()
	_, err = verifier.ValidatePopToken(expiredToken)
	assert.NotNilf(t, err, "PoP verification succeed.")
	assert.Containsf(t, err.Error(), "Token is expired", "Error message is not as expected")

	// 'ts' claim is not present in the payload
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetKid(tsClaimsMissing).GetToken() //(time.Now().Unix(), "", tsClaimsMissing)
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'ts' claim is missing")

	// 'ts' claim in the payload is of unknown type and cannot be parsed. 'ts' is set to the default timestamp which has expired
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(tsClaimsTypeUnknown).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.Containsf(t, err.Error(), "Token is expired", "Error message is not as expected")

	// Request and validation for the PoP token are running for different hostnames
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("wrongHostnme").GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid Pop token due to host mismatch. Expected: \"testHostname\", received: \"wrongHostnme\"")

	// 'cnf' is not present in the access token
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(atCnfClaimMissing).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "could not retrieve 'cnf' claim from access token")

	// 'cnf' in the access token does not have the right value
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(atCnfClaimWrong).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "PoP token validate failed: 'cnf' claim mismatch")

	// 'cnf' claim in the payload is not present
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(cnfClaimsMissing).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'cnf' claim is missing")

	// 'cnf' claim in the payload does not have 'jwk'
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(cnfJwkClaimsMissing).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'cnf' claim is not in expected format")

	// 'jwk' in 'cnf' claim in the payload is empty
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(cnfJwkClaimsEmpty).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'jwk' claim is empty")

	// 'jwk' in the 'cnf' claim in the payload is not of string type
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(cnfJwkClaimsWrong).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.ErrorContains(t, err, "failed while parsing 'jwk' claim in PoP token")

	// 'jwk' in the 'cnf' claim in the access token is not of string type
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(accessTokenCnfWrong).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.ErrorContains(t, err, "failed while parsing 'cnf' in access token")

	// 'u' claim is not present in the payload
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetKid(uClaimsMissing).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'u' claim is missing")

	// 'u' claim in the payload is not of string type
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(uClaimsWrongType).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'u' claim should be of string")

	// 'at' claim in the payload is not of type string
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(atClaimsWrongType).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. 'at' claim should be string")

	// 'at' claim is not present in the payload
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(atClaimsMissing).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. access token missing")

	// 'at' claim value in they payload is not correct and could not be parsed
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(atClaimIncorrect).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.ErrorContains(t, err, "could not parse access token in PoP token")

	// RSA verify error due to invalid signature in the PoP token
	invalidToken, _ = NewPoPTokenBuilder().SetTimestamp(time.Now().Unix()).SetHostName("testHostname").SetKid(signatureWrongType).GetToken()
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.ErrorContains(t, err, "RSA verify err: crypto/rsa: verification error")
}
