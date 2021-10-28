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
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
)

const (
	popAccessToken = `{ "aud": "client", "iss" : "kd", "exp" : "%d","cnf": {"kid":"%s","xms_ksl":"sw"} }`
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
	var signerOpts = jose.SignerOptions{}
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
	cnfClaimsMissing    = "cnfClaimsMissing"
	cnfJwkClaimsWrong   = "cnfJwkClaimsWrong"
)

func GeneratePoPToken(ts int64, hostName, kid string) (string, error) {
	if strings.Contains(kid, badTokenKey) {
		return kid, nil
	}
	popKey, err := NewSWPoPKey()
	if err != nil {
		return "", fmt.Errorf("Failed to generate Pop key. Error:%+v", err)
	}

	key, err := NewSwkKey()
	if err != nil {
		return "", fmt.Errorf("Failed to generate SF key. Error:%+v", err)
	}

	var cnf string = kid
	if cnf == "" {
		cnf = popKey.KeyID()
	}

	at, err := key.GenerateToken([]byte(fmt.Sprintf(popAccessToken, time.Now().Add(time.Minute*5).Unix(), cnf)))
	if err != nil {
		return "", fmt.Errorf("Error when generating token. Error:%+v", err)
	}
	var header, headerB64 string
	nonce := uuid.New().String()
	nonce = strings.Replace(nonce, "-", "", -1)
	keyID := popKey.KeyID()
	algo := popKey.Alg()
	typ := "pop"
	if kid == headerBadKeyID {
		keyID = ""
	}
	if kid == headerBadAlgo {
		algo = "wrong"
	}
	if kid == headerBadtyp {
		typ = "wrong"
	}
	header = fmt.Sprintf(`{"typ":"%s","alg":"%s","kid":"%s"}`, typ, algo, keyID)

	if kid == headerBadtypType {
		wrongTyp := 1
		header = fmt.Sprintf(`{"typ":"%d","alg":"%s","kid":"%s"}`, wrongTyp, algo, keyID)
	}

	if kid == headerBadtypMissing {
		header = fmt.Sprintf(`{"alg":"%s","kid":"%s"}`, algo, keyID)
	}

	headerB64 = base64.RawURLEncoding.EncodeToString([]byte(header))

	var payload, payloadB64 string
	payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, ts, hostName, popKey.Jwk(), nonce)
	if kid == tsClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "u": "%d", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, 1, popKey.Jwk(), nonce)
	}
	if kid == uClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "cnf":{"jwk":%s}, "nonce":"%s"}`, at, ts, popKey.Jwk(), nonce)
	}
	if kid == cnfClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s"`, at, ts, hostName)
	}
	if kid == cnfJwkClaimsWrong {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{}, "nonce":"%s"}`, at, ts, hostName, nonce)
	}

	payloadB64 = base64.RawURLEncoding.EncodeToString([]byte(payload))
	h256 := sha256.Sum256([]byte(headerB64 + "." + payloadB64))
	signature, err := popKey.Sign(h256[:])
	if err != nil {
		return "", fmt.Errorf("Error while signing pop key. Error:%+v", err)
	}
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	finalToken := headerB64 + "." + payloadB64 + "." + signatureB64
	return finalToken, nil
}

func TestPopTokenVerifier_Verify(t *testing.T) {
	verifier := NewPoPVerifier("testHostname", 15*time.Minute)

	validToken, _ := GeneratePoPToken(time.Now().Unix(), "testHostname", "")
	_, err := verifier.ValidatePopToken(validToken)
	assert.NoError(t, err)

	invalidToken, _ := GeneratePoPToken(time.Now().Unix(), "", badTokenKey)
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "PoP token invalid schema. Token length: 1")

	invalidToken, _ = GeneratePoPToken(time.Now().Unix(), "testHostname", fmt.Sprintf("%s.%s.%s", badTokenKey, badTokenKey, badTokenKey))
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Could not parse PoP token. Error: invalid character 'm' looking for beginning of value")

	invalidToken, _ = GeneratePoPToken(time.Now().Unix(), "", headerBadKeyID)
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "No KeyID found in PoP token header")

	invalidToken, _ = GeneratePoPToken(time.Now().Unix(), "", headerBadAlgo)
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Wrong algorithm found in PoP token header")

	invalidToken, _ = GeneratePoPToken(time.Now().Unix(), "", headerBadtyp)
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Wrong typ of token. Expected pop token")

	invalidToken, _ = GeneratePoPToken(time.Now().Unix(), "", headerBadtypMissing)
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. Typ claim missing")

	expiredToken, _ := GeneratePoPToken(time.Now().Add(time.Minute*-20).Unix(), "", "")
	_, err = verifier.ValidatePopToken(expiredToken)
	assert.NotNilf(t, err, "PoP verification succeed.")

	invalidToken, _ = GeneratePoPToken(time.Now().Unix(), "", tsClaimsMissing)
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. ts claim missing")

	invalidToken, _ = GeneratePoPToken(time.Now().Unix(), "wrongHostnme", "")
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid Pop token. Host mismatch. Expected: testHostname, Req received: wrongHostnme")

	invalidToken, _ = GeneratePoPToken(time.Now().Unix(), "", uClaimsMissing)
	_, err = verifier.ValidatePopToken(invalidToken)
	assert.EqualError(t, err, "Invalid token. u claim missing")
}
