package lib

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestCheckLdap(t *testing.T) {
	// test 1
	// disabled anonymous access
	opts := LDAPOptions{
		ServerAddress:        "localhost",
		ServerPort:           "10389",
		BindDN:               "uid=admin,ou=system",
		BindPassword:         "secret",
		UserSearchDN:         "o=Company",
		UserSearchFilter:     DefaultUserSearchFilter,
		UserAttribute:        DefaultUserAttribute,
		GroupSearchDN:        "o=Company",
		GroupSearchFilter:    DefaultGroupSearchFilter,
		GroupMemberAttribute: DefaultGroupMemberAttribute,
		GroupNameAttribute:   DefaultGroupNameAttribute,
		SkipTLSVerification:  true,
		StartTLS:             true,
	}
	s := Server{LDAP: opts}
	resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid:12345")))
	if status != http.StatusOK {
		t.Error(resp.Status.Error)
	}
	var (
		testFnd, adminFnd bool
	)
	for _, g := range resp.Status.User.Groups {
		if g == "test" {
			testFnd = true
		}

		if g == "guard-test" {
			adminFnd = true
		}
	}
	if !testFnd || !adminFnd {
		t.Errorf(`expected: group list ["test","guard-test"], got %s`, strings.Join(resp.Status.User.Groups, ","))
	}
	fmt.Print(resp.Status)
}

func TestParseEncodedToken(t *testing.T) {
	user, pass, ok := parseEncodedToken(base64.StdEncoding.EncodeToString([]byte("user1:12345")))
	if !ok {
		t.Error("Expected: parsing successfull, got parsing unsuccessfull")
	}
	if user != "user1" {
		t.Error("Expected: user: user1, got user:", user)
	}
	if pass != "12345" {
		t.Error("Expected: password: 12345, got password:", pass)
	}
}
