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
	"time"

	"github.com/google/uuid"
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
	BadTokenKey         = "badToken"
	HeaderBadKeyID      = "headerBadKeyID"
	HeaderBadAlgo       = "headerBadAlgo"
	HeaderBadtyp        = "headerBadtyp"
	HeaderBadtypType    = "headerBadtypType"
	HeaderBadtypMissing = "headerBadtypMissing"
	UClaimsMissing      = "uClaimsMissing"
	TsClaimsMissing     = "tsClaimsMissing"
	AtClaimsMissing     = "atClaimsMissing"
	AtClaimIncorrect    = "atClaimIncorrect"
	CnfClaimsMissing    = "cnfClaimsMissing"
	CnfJwkClaimsEmpty   = "cnfJwkClaimsEmpty"
	CnfJwkClaimsWrong   = "cnfJwkClaimsWrong"
	CnfJwkClaimsMissing = "cnfJwkClaimsMissing"
	AccessTokenCnfWrong = "accessTokenCnfWrong"
	AtClaimsWrongType   = "atClaimsWrongType"
	AtCnfClaimMissing   = "atCnfClaimMissing"
	AtCnfClaimWrong     = "atCnfClaimWrong"
	TsClaimsTypeString  = "tsClaimsTypeString"
	TsClaimsTypeUnknown = "tsClaimsTypeUnknown"
	UClaimsWrongType    = "uClaimsWrongType"
	SignatureWrongType  = "signatureWrongType"
	NonceClaimMissing   = "nonceMissing"
	NonceClaimHardcoded = "nonceHardcoded"
	NonceClaimNotString = "nonceNotString"
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
	nonce    string
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
	if b.kid == HeaderBadKeyID {
		keyID = ""
	}
	if b.kid == HeaderBadAlgo {
		algo = "wrong"
	}
	if b.kid == HeaderBadtyp {
		typ = "wrong"
	}
	header := fmt.Sprintf(`{"typ":"%s","alg":"%s","kid":"%s"}`, typ, algo, keyID)

	if b.kid == HeaderBadtypType {
		wrongTyp := 1
		header = fmt.Sprintf(`{"typ":%d,"alg":"%s","kid":"%s"}`, wrongTyp, algo, keyID)
	}
	if b.kid == HeaderBadtypMissing {
		header = fmt.Sprintf(`{"alg":"%s","kid":"%s"}`, algo, keyID)
	}
	b.token.Header = base64.RawURLEncoding.EncodeToString([]byte(header))
	return nil
}

// A method that sets the payload of the PoP token
func (b *PoPTokenBuilderImpl) SetPayload() error {
	var cnf string
	if b.kid == AtCnfClaimWrong {
		cnf = "wrongCnf"
	} else {
		cnf = b.popKey.KeyID()
	}

	if b.kid == NonceClaimHardcoded {
		b.nonce = "hardcodedNonce"
	} else {
		b.nonce = strings.Replace(uuid.New().String(), "-", "", -1)
	}

	accessTokenData := fmt.Sprintf(popAccessToken, time.Now().Add(time.Minute*5).Unix(), cnf)
	if b.kid == AtCnfClaimMissing {
		accessTokenData = fmt.Sprintf(popAccessTokenWithoutCnf, time.Now().Add(time.Minute*5).Unix())
	}
	if b.kid == AccessTokenCnfWrong {
		accessTokenData = fmt.Sprintf(`{ "aud": "client", "iss" : "kd", "exp" : "%d","cnf": {"kid":1} }`, time.Now().Add(time.Minute*5).Unix())
	}

	at, err := b.swkKey.GenerateToken([]byte(accessTokenData))
	if err != nil {
		return fmt.Errorf("error when generating token. Error:%+v", err)
	}

	payload := fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, b.ts, b.hostName, b.popKey.Jwk(), b.nonce)
	if b.kid == TsClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "u": "%d", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, 1, b.popKey.Jwk(), b.nonce)
	}
	if b.kid == UClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "cnf":{"jwk":%s}, "nonce":"%s"}`, at, b.ts, b.popKey.Jwk(), b.nonce)
	}
	if b.kid == CnfJwkClaimsEmpty {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{}, "nonce":"%s"}`, at, b.ts, b.hostName, b.nonce)
	}
	if b.kid == CnfJwkClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":1, "nonce":"%s"}`, at, b.ts, b.hostName, b.nonce)
	}
	if b.kid == CnfJwkClaimsWrong {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":1}, "nonce":"%s"}`, at, b.ts, b.hostName, b.nonce)
	}
	if b.kid == CnfClaimsMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "nonce": "%s"}`, at, b.ts, b.hostName, b.nonce)
	}
	if b.kid == TsClaimsTypeString {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : "%s", "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, strconv.FormatInt(b.ts, 10), b.hostName, b.popKey.Jwk(), b.nonce)
	}
	if b.kid == TsClaimsTypeUnknown {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %t, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, at, bool(true), b.hostName, b.popKey.Jwk(), b.nonce)
	}
	if b.kid == AtClaimsWrongType {
		payload = fmt.Sprintf(`{ "at" : %d, "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, 12, b.ts, b.hostName, b.popKey.Jwk(), b.nonce)
	}
	if b.kid == UClaimsWrongType {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": %d, "cnf":{"jwk":%s}, "nonce":"%s"}`, at, b.ts, 1, b.popKey.Jwk(), b.nonce)
	}
	if b.kid == AtClaimsMissing {
		payload = fmt.Sprintf(`{ "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, b.ts, b.hostName, b.popKey.Jwk(), b.nonce)
	}
	if b.kid == AtClaimIncorrect {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":"%s"}`, fmt.Sprintf("%s.%s.%s", BadTokenKey, BadTokenKey, BadTokenKey), b.ts, b.hostName, b.popKey.Jwk(), b.nonce)
	}
	if b.kid == NonceClaimMissing {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":%s}}`, at, b.ts, b.hostName, b.popKey.Jwk())
	}
	if b.kid == NonceClaimNotString {
		payload = fmt.Sprintf(`{ "at" : "%s", "ts" : %d, "u": "%s", "cnf":{"jwk":%s}, "nonce":%d}`, at, b.ts, b.hostName, b.popKey.Jwk(), 1)
	}
	b.token.Payload = base64.RawURLEncoding.EncodeToString([]byte(payload))
	return nil
}

// A method that sets the signature of the PoP token
func (b *PoPTokenBuilderImpl) SetSignature() error {
	if b.kid == SignatureWrongType {
		b.token.Signature = "wrongSignature"
		return nil
	}
	h256 := sha256.Sum256([]byte(b.token.Header + "." + b.token.Payload))
	signature, err := b.popKey.Sign(h256[:])
	if err != nil {
		return fmt.Errorf("error while signing pop key. Error:%+v", err)
	}
	b.token.Signature = base64.RawURLEncoding.EncodeToString(signature)
	return nil
}

// A method that returns the final PoP token as a string
func (b *PoPTokenBuilderImpl) GetToken() (string, error) {
	var err error
	b.popKey, err = NewSWPoPKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate Pop key. Error:%+v", err)
	}
	b.swkKey, err = NewSwkKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate SF key. Error:%+v", err)
	}

	if strings.Contains(b.kid, BadTokenKey) {
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
