package ldap

import (
	"crypto/tls"
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

// getTLSconfig returns a tls configuration used
// to build a TLSlistener for TLS or StartTLS
func (s *ldapServer) getTLSconfig() (*tls.Config, error) {
	srvCert, srvKey, err := s.certStore.NewServerCertPairBytes(cert.AltNames{
		DNSNames: []string{"server"},
		IPs:      []net.IP{net.ParseIP(serverAddr)},
	})
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(srvCert, srvKey)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionSSL30,
		MaxVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		ServerName:   serverAddr,
	}, nil
}

func ldapServerSetup(secureConn bool, userSearchDN, groupSearchDN string) (*ldapServer, error) {
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
	if bindDN == "uid=admin,ou=system" && bindSimplePw == "secret" {
		return ldapserver.LDAPResultSuccess, nil
	}

	// for userDN
	if bindDN == "uid=nahid,ou=users,o=Company" && bindSimplePw == "secret" {
		return ldapserver.LDAPResultSuccess, nil
	}

	// for userDN
	if bindDN == "uid=shuvo,ou=users,o=Company" && bindSimplePw == "secret" {
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
			"uid=nahid,ou=users,o=Company",
			[]*ldapserver.EntryAttribute{
				{"cn", []string{"nahid"}},
			},
		})
	}

	// one entry
	if searchReq.Filter == "(&(objectClass=person)(uid=shuvo))" {
		entries = append(entries, &ldapserver.Entry{
			"uid=shuvo,ou=users,o=Company",
			[]*ldapserver.EntryAttribute{
				{"cn", []string{"shuvo"}},
			},
		})
	}

	// multiple entry
	if searchReq.Filter == "(&(objectClass=person)(id=nahid))" {
		entries = append(entries, &ldapserver.Entry{
			"uid=nahid,ou=users,o=Company",
			[]*ldapserver.EntryAttribute{
				{"cn", []string{"nahid"}},
				{"id", []string{"1204"}},
			},
		}, &ldapserver.Entry{
			"uid=shuvo,ou=users,o=Company",
			[]*ldapserver.EntryAttribute{
				{"cn", []string{"shuvo"}},
				{"id", []string{"1204"}},
			},
		})
	}

	// one entry
	if searchReq.Filter == "(&(objectClass=groupOfNames)(member=uid=nahid,ou=users,o=Company))" {
		entries = append(entries, &ldapserver.Entry{
			"id=1,ou=groups,o=Company",
			[]*ldapserver.EntryAttribute{
				{"cn", []string{"group1"}},
			},
		}, &ldapserver.Entry{
			"id=1,ou=groups,o=Company",
			[]*ldapserver.EntryAttribute{
				{"cn", []string{"group2"}},
			},
		})
	}
	return ldapserver.ServerSearchResult{entries, []string{}, []ldapserver.Control{}, ldapserver.LDAPResultSuccess}, nil
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
	srv, err := ldapServerSetup(secureConn, "o=Company,ou=users", "o=Company,ou=groups")
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
