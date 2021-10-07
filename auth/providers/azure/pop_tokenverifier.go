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
	"bytes"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gopkg.in/square/go-jose.v2/jwt"
	"k8s.io/klog/v2"

	"github.com/pkg/errors"
)

// PopTokenVerifier is validator for PoP tokens.
type PopTokenVerifier struct {
	hostName  string
	ValidTill time.Duration
}

func NewPoPVerifier(hostName string, validTill time.Duration) *PopTokenVerifier {
	return &PopTokenVerifier{
		ValidTill: validTill,
		hostName:  hostName,
	}
}

// Claims maintains token claims
type Claims map[string]interface{}

const (
	//TypPoP signifies pop token
	TypPoP = "pop"
	//TypJWT signifies AAD JWT token
	TypJWT = "JWT"
	//AlgoRS256 signifies signing algorithm
	AlgoRS256 = "RS256"
)

// Jwk maintains public key info
type Jwk struct {
	E   string `json:"e"`
	Kty string `json:"kty"`
	N   string `json:"n"`
}

// Acquire is validating the pop token
func (p *PopTokenVerifier) ValidatePopToken(token string) (string, error) {
	data := strings.Split(token, ".")
	if len(data) != 3 || data[0] == "" || data[1] == "" || data[2] == "" {
		return "", errors.Errorf("PoP token invalid schema. Token length: %d", len(data))
	}

	ptoken, err := jwt.ParseSigned(token)
	if err != nil {
		return "", errors.Errorf("Could not parse PoP token. Error: %+v", err)
	}
	var claims Claims
	_ = ptoken.UnsafeClaimsWithoutVerification(&claims)

	// This can never happens since the first 'if len(date) != 3' check if the header is present
	if len(ptoken.Headers) <= 0 {
		return "", errors.Errorf("No header found in PoP token")
	}

	if ptoken.Headers[0].KeyID == "" {
		return "", errors.Errorf("No KeyID found in PoP token header")
	}

	if ptoken.Headers[0].Algorithm != AlgoRS256 {
		return "", errors.Errorf("Wrong algorithm found in PoP token header")
	}

	if typ, ok := ptoken.Headers[0].ExtraHeaders["typ"]; ok {
		if tokenType, ok := typ.(string); ok {
			if !strings.EqualFold(tokenType, TypPoP) {
				return "", errors.Errorf("Wrong typ of token. Expected pop token")
			}
		} else {
			return "", errors.Errorf("Invalid token. Typ claim should be string")
		}
	} else {
		return "", errors.Errorf("Invalid token. Typ claim missing")
	}

	/* Verify expiry time */
	// This is useful for fail fast.
	now := time.Now()
	var issuedTime time.Time
	if ts, ok := claims["ts"]; ok {
		convertTime(ts, &issuedTime)
		expireat := issuedTime.Add(p.ValidTill * time.Minute)
		if expireat.Before(now) {
			return "", errors.Errorf("Token is expired. Now: %v, Valid till: %v", now, expireat)
		}
	} else {
		return "", errors.Errorf("Invalid token. ts claim missing")
	}

	/* Verify host */
	if uc, ok := claims["u"]; ok {
		if reqHostName, ok := uc.(string); ok {
			if klog.V(10).Enabled() {
				klog.V(10).Infoln("pop token validation running with hostName: %s. Request is coming for hostName: %s", p.hostName, reqHostName)
			}
			if !strings.EqualFold(reqHostName, p.hostName) {
				return "", errors.Errorf("Invalid Pop token. Host mismatch. Expected: %s, Req received: %s", p.hostName, reqHostName)
			}
		} else {
			return "", errors.Errorf("Invalid token. u claim should be string")
		}
	} else {
		return "", errors.Errorf("Invalid token. u claim missing")
	}

	var cnf map[string]interface{}
	if cnfclaim, ok := claims["cnf"]; ok {
		if cnf, ok = cnfclaim.(map[string]interface{}); !ok {
			return "", errors.Errorf("Invalid token. Cnf claim missing")
		}
	}

	var jwk Jwk
	if err := marshalGenericTo(cnf["jwk"], &jwk); err != nil {
		return "", errors.Errorf("failed while parsing jwk claim in PoP token : %v", err)
	}

	/* Verify signing of PoP token */
	message := fmt.Sprintf("%s.%s", data[0], data[1])

	signature, err := base64.RawURLEncoding.DecodeString(data[2])
	if err != nil {
		return "", errors.Errorf("Failed to decode signed message with url decoding .Error: %+v", err)
	}

	n, _ := base64.RawURLEncoding.DecodeString(jwk.N)
	e, _ := base64.RawURLEncoding.DecodeString(jwk.E)
	z := new(big.Int)
	z.SetBytes(n)

	var buffer bytes.Buffer
	buffer.WriteByte(0)
	buffer.Write(e)
	exponent := binary.BigEndian.Uint32(buffer.Bytes())
	publicKey := &rsa.PublicKey{N: z, E: int(exponent)}

	hasher := crypto.SHA256.New()
	_, err = hasher.Write([]byte(message))
	if err != nil {
		return "", errors.Errorf("Failed to write message to hasher. Error:%+v", err)
	}
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hasher.Sum(nil), signature)

	if err != nil {
		return "", errors.Errorf("RSA verify err: %+v", err)
	}

	return claims["at"].(string), nil
}

func convertTime(i interface{}, tm *time.Time) {
	switch iat := i.(type) {
	case float64:
		*tm = time.Unix(int64(iat), 0)
	case int64:
		*tm = time.Unix(iat, 0)
	case string:
		v, _ := strconv.ParseInt(iat, 10, 64)
		*tm = time.Unix(v, 0)
	}
}
