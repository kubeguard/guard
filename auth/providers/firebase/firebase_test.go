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
package firebase

import (
	"context"
	"testing"

	fbauth "firebase.google.com/go/auth"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
)

const (
	validUID    = "uid1234567890"
	invalidUser = "invalidUser"
	email       = "user@test.com"
)

var (
	validToken = &fbauth.Token{
		UID: validUID,
	}
	validUser = &fbauth.UserRecord{
		UserInfo: &fbauth.UserInfo{
			Email: email,
		},
	}
)

func TestValidToken(t *testing.T) {
	expectedUser := &authv1.UserInfo{
		Username: email,
		UID:      validUID,
	}

	g := NewGomegaWithT(t)
	client := &Authenticator{newTestFirebaseAuthClient()}

	user, err := client.Check(validUID)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(user).NotTo(BeNil())
	g.Expect(user).To(Equal(expectedUser))
}

func TestInValidToken(t *testing.T) {
	g := NewGomegaWithT(t)
	client := &Authenticator{newTestFirebaseAuthClient()}

	user, err := client.Check("bogus token")
	g.Expect(err).To(HaveOccurred())
	g.Expect(user).To(BeNil())
}

func TestTokenWithInvalidUser(t *testing.T) {
	g := NewGomegaWithT(t)
	client := &Authenticator{newTestFirebaseAuthClient()}

	user, err := client.Check(invalidUser)
	g.Expect(err).To(HaveOccurred())
	g.Expect(user).To(BeNil())
}

// Firebase client mock
type testFirebaseAuthClient struct {
}

func newTestFirebaseAuthClient() FirebaseAuth {
	return &testFirebaseAuthClient{}
}

func (c *testFirebaseAuthClient) VerifyIDTokenAndCheckRevoked(context context.Context, idToken string) (*fbauth.Token, error) {
	// user exists and has valid token
	if idToken == validUID {
		return validToken, nil
	}
	// user no longer exists, but token is still valid (until its EOL)
	if idToken == invalidUser {
		return &fbauth.Token{}, nil
	}
	return nil, errors.New("invalid token")
}

func (c *testFirebaseAuthClient) GetUser(context context.Context, uid string) (*fbauth.UserRecord, error) {
	if uid == validUID {
		return validUser, nil
	}
	return nil, errors.New("invalid user ID")
}
