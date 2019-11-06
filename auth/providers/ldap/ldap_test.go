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
package ldap

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/golang/glog"
	ldapserver "github.com/nmcclain/ldap"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
)

const (
	serverAddr   = "127.0.0.1"
	inSecurePort = "8089"
	securePort   = "8889"
)

type ldapServer struct {
	server     *ldapserver.Server
	secureConn bool
	certStore  *certstore.CertStore
}

func (s *ldapServer) start() {
	if s.secureConn {
		srvCert, srvKey, err := s.certStore.NewServerCertPairBytes(cert.AltNames{
			DNSNames: []string{"server"},
			IPs:      []net.IP{net.ParseIP(serverAddr)},
		})
		if err != nil {
			glog.Fatal(err)
		}

		err = s.certStore.WriteBytes("srv", srvCert, srvKey)
		if err != nil {
			glog.Fatal(err)
		}

		if err := s.server.ListenAndServeTLS(serverAddr+":"+securePort, s.certStore.CertFile("srv"), s.certStore.KeyFile("srv")); err != nil {
			glog.Fatal(err)
		}
	} else {
		if err := s.server.ListenAndServe(serverAddr + ":" + inSecurePort); err != nil {
			glog.Fatal(err)
		}
	}
}

func (s *ldapServer) stop() {

}

func ldapServerSetup(secureConn bool) (*ldapServer, error) {
	//Create a new LDAP Server
	server := ldapserver.NewServer()
	handler := ldapHandler{}

	server.BindFunc("", handler)
	server.SearchFunc("", handler)

	srv := &ldapServer{
		server:     server,
		secureConn: secureConn,
	}

	if secureConn {
		store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(os.TempDir(), "ldap-certs"), "test")
		if err != nil {
			return nil, err
		}

		err = store.InitCA()
		if err != nil {
			return nil, err
		}
		srv.certStore = store
	}

	return srv, nil
}

type ldapHandler struct {
}

func (h ldapHandler) Bind(bindDN, bindSimplePw string, conn net.Conn) (ldapserver.LDAPResultCode, error) {
	fmt.Println("*********bind**************")
	fmt.Println(bindDN, bindSimplePw)
	fmt.Println("*********bind-end**************")

	const bindSimplePwSecret = "secret"
	if bindDN == "uid=admin,ou=system" && bindSimplePw == bindSimplePwSecret {
		return ldapserver.LDAPResultSuccess, nil
	}

	// for userDN
	if bindDN == "uid=nahid,ou=users,o=Company" && bindSimplePw == bindSimplePwSecret {
		return ldapserver.LDAPResultSuccess, nil
	}

	// for userDN
	if bindDN == "uid=shuvo,ou=users,o=Company" && bindSimplePw == bindSimplePwSecret {
		return ldapserver.LDAPResultSuccess, nil
	}
	if bindDN == "" && bindSimplePw == "" {
		return ldapserver.LDAPResultSuccess, nil
	}
	return ldapserver.LDAPResultInvalidCredentials, nil
}

func (h ldapHandler) Search(boundDN string, searchReq ldapserver.SearchRequest, conn net.Conn) (ldapserver.ServerSearchResult, error) {
	fmt.Println("*********search**************")
	fmt.Println(boundDN)
	fmt.Println(searchReq)

	var entries []*ldapserver.Entry

	// one entry
	if searchReq.Filter == "(&(objectClass=person)(uid=nahid))" {
		entries = append(entries, &ldapserver.Entry{
			DN: "uid=nahid,ou=users,o=Company",
			Attributes: []*ldapserver.EntryAttribute{
				{Name: "cn", Values: []string{"nahid"}},
			},
		})
	}

	// one entry
	if searchReq.Filter == "(&(objectClass=person)(uid=shuvo))" {
		entries = append(entries, &ldapserver.Entry{
			DN: "uid=shuvo,ou=users,o=Company",
			Attributes: []*ldapserver.EntryAttribute{
				{Name: "cn", Values: []string{"shuvo"}},
			},
		})
	}

	// multiple entry
	if searchReq.Filter == "(&(objectClass=person)(id=nahid))" {
		entries = append(entries, &ldapserver.Entry{
			DN: "uid=nahid,ou=users,o=Company",
			Attributes: []*ldapserver.EntryAttribute{
				{Name: "cn", Values: []string{"nahid"}},
				{Name: "id", Values: []string{"1204"}},
			},
		}, &ldapserver.Entry{
			DN: "uid=shuvo,ou=users,o=Company",
			Attributes: []*ldapserver.EntryAttribute{
				{Name: "cn", Values: []string{"shuvo"}},
				{Name: "id", Values: []string{"1204"}},
			},
		})
	}

	// one entry
	if searchReq.Filter == "(&(objectClass=groupOfNames)(member=uid=nahid,ou=users,o=Company))" {
		entries = append(entries, &ldapserver.Entry{
			DN: "id=1,ou=groups,o=Company",
			Attributes: []*ldapserver.EntryAttribute{
				{Name: "cn", Values: []string{"group1"}},
			},
		}, &ldapserver.Entry{
			DN: "id=1,ou=groups,o=Company",
			Attributes: []*ldapserver.EntryAttribute{
				{Name: "cn", Values: []string{"group2"}},
			},
		})
	}
	return ldapserver.ServerSearchResult{Entries: entries, Referrals: []string{}, Controls: []ldapserver.Control{}, ResultCode: ldapserver.LDAPResultSuccess}, nil
}

func TestCheckLdapInSecure(t *testing.T) {

	opts := Options{
		ServerAddress:        serverAddr,
		ServerPort:           inSecurePort,
		BindDN:               "uid=admin,ou=system",
		BindPassword:         "secret",
		UserSearchDN:         "o=Company,ou=users",
		UserSearchFilter:     DefaultUserSearchFilter,
		UserAttribute:        DefaultUserAttribute,
		GroupSearchDN:        "o=Company,ou=groups",
		GroupSearchFilter:    DefaultGroupSearchFilter,
		GroupMemberAttribute: DefaultGroupMemberAttribute,
		GroupNameAttribute:   DefaultGroupNameAttribute,
		SkipTLSVerification:  true,
		StartTLS:             false,
		IsSecureLDAP:         false,
	}
	s := Authenticator{
		opts: opts,
	}

	runTest(t, false, s, "Insecure LDAP")
}

func TestCheckLdapSecure(t *testing.T) {
	opts := Options{
		ServerAddress:        serverAddr,
		ServerPort:           securePort,
		BindDN:               "uid=admin,ou=system",
		BindPassword:         "secret",
		UserSearchDN:         "o=Company,ou=users",
		UserSearchFilter:     DefaultUserSearchFilter,
		UserAttribute:        DefaultUserAttribute,
		GroupSearchDN:        "o=Company,ou=groups",
		GroupSearchFilter:    DefaultGroupSearchFilter,
		GroupMemberAttribute: DefaultGroupMemberAttribute,
		GroupNameAttribute:   DefaultGroupNameAttribute,
		SkipTLSVerification:  false,
		StartTLS:             false,
		IsSecureLDAP:         true,
	}
	s := Authenticator{
		opts: opts,
	}

	runTest(t, true, s, "Secure LDAP")
}

func runTest(t *testing.T, secureConn bool, s Authenticator, serverType string) {
	srv, err := ldapServerSetup(secureConn)
	if err != nil {
		t.Fatal(err)
	}

	go srv.start()
	defer srv.stop()
	// wait for server to start
	time.Sleep(10 * time.Second)

	if secureConn {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(srv.certStore.CACertBytes())
		s.opts.CaCertFile = srv.certStore.CertFile("ca")
		s.opts.CaCertPool = caCertPool
	}

	dataset := []struct {
		testName      string
		token         string
		authenticated bool
		username      string
		groups        []string
		userAttribute string
	}{
		{
			"authentication successful",
			"nahid:secret",
			true,
			"nahid",
			[]string{"group1", "group2"},
			DefaultUserAttribute,
		},
		{
			"authentication unsuccessful, reason multiple entry when searching userDN",
			"nahid:secret",
			false,
			"",
			nil,
			"id",
		},
		{
			"authentication unsuccessful, reason empty entry when searching userDN",
			"nahid1:secret",
			false,
			"",
			nil,
			DefaultUserAttribute,
		},
		{
			"authentication unsuccessful, reason invalid token",
			"invalid_token",
			false,
			"",
			nil,
			DefaultUserAttribute,
		},
		{
			"authentication unsuccessful, wrong username or password",
			"nahid:12345",
			false,
			"",
			nil,
			DefaultUserAttribute,
		},
		{
			"authentication successful, empty group",
			"shuvo:secret",
			true,
			"shuvo",
			[]string{},
			DefaultUserAttribute,
		},
	}

	// This Run will not return until the parallel tests finish.
	t.Run("ldap", func(t *testing.T) {
		for _, tc := range dataset {
			t.Run(serverType+": "+tc.testName, func(t *testing.T) {
				t.Log(tc)

				serv := s
				serv.opts.UserAttribute = tc.userAttribute

				// set up client token
				token := base64.StdEncoding.EncodeToString([]byte(tc.token))

				resp, err := serv.Check(token)
				if tc.authenticated {
					if assert.Nil(t, err) {
						if resp.Username != tc.username {
							t.Errorf("Expected username %v, got %v", tc.username, resp.Username)
						}
						if len(resp.Groups) != len(tc.groups) {
							t.Errorf("Expected group size %v, got %v", len(tc.groups), len(resp.Groups))
						} else {
							if len(resp.Groups) > 0 {
								if !reflect.DeepEqual(resp.Groups, tc.groups) {
									t.Errorf("Expected groups %v, got %v", tc.groups, resp.Groups)
								}
							}
						}
					}
				} else {
					assert.NotNil(t, err)
					assert.Nil(t, resp)
				}
			})
		}
	})
}

func TestParseEncodedToken(t *testing.T) {
	user, pass, ok := parseEncodedToken(base64.StdEncoding.EncodeToString([]byte("user1:12345")))
	if !ok {
		t.Error("Expected: parsing successful, got parsing unsuccessful")
	}
	if user != "user1" {
		t.Error("Expected: user: user1, got user:", user)
	}
	if pass != "12345" {
		t.Error("Expected: password: 12345, got password:", pass)
	}
}
