package ldap

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"log"
	"net"
	"os"
	"reflect"
	"testing"
	"time"
	"path/filepath"

	"github.com/appscode/kutil/tools/certstore"
	"github.com/go-ldap/ldap"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/vjeantet/ldapserver"
	"k8s.io/client-go/util/cert"
)

const (
	serverAddr   = "127.0.0.1"
	inSecurePort = "8089"
	securePort   = "8889"
)

type ldapServer struct {
	server     *ldapserver.Server
	secureConn bool
	stopCh     chan bool
	certStore  *certstore.CertStore
}

func (s *ldapServer) start() {
	var err error
	go func() {
		if s.secureConn {
			tlsConfig, err := s.getTLSconfig()
			if err == nil {
				err = s.server.ListenAndServe(serverAddr+":"+securePort, func(s *ldapserver.Server) {
					s.Listener = tls.NewListener(s.Listener, tlsConfig)
				})
			}
		} else {
			err = s.server.ListenAndServe(serverAddr + ":" + inSecurePort)
		}
		log.Println("LDAP Server: ", err)
	}()

	<-s.stopCh
	close(s.stopCh)
	s.server.Stop()
}

func (s *ldapServer) stop() {
	s.stopCh <- true
	if s.certStore != nil {
		os.RemoveAll(s.certStore.Location())
	}
}

// getTLSconfig returns a tls configuration used
// to build a TLSlistener for TLS or StartTLS
func (s *ldapServer) getTLSconfig() (*tls.Config, error) {
	srvCert, srvKey, err := s.certStore.NewServerCertPair("server", cert.AltNames{IPs: []net.IP{net.ParseIP(serverAddr)}})
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

	routes := ldapserver.NewRouteMux()

	routes.Bind(handleBind).AuthenticationChoice("simple")

	routes.Search(handleUserSearch).BaseDn(userSearchDN)

	routes.Search(handleGroupSearch).BaseDn(groupSearchDN)

	server.Handle(routes)

	srv := &ldapServer{
		server:     server,
		stopCh:     make(chan bool),
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

// handleBind return Success if username and password matched
func handleBind(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	r := m.GetBindRequest()
	res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)

	log.Println("Bind :", r.Name(), r.AuthenticationSimple())

	// for baseDN
	if string(r.Name()) == "uid=admin,ou=system" && string(r.AuthenticationSimple()) == "secret" {
		w.Write(res)
		return
	}

	// for userDN
	if string(r.Name()) == "uid=nahid,ou=users,o=Company" && string(r.AuthenticationSimple()) == "secret" {
		w.Write(res)
		return
	}

	// for userDN
	if string(r.Name()) == "uid=shuvo,ou=users,o=Company" && string(r.AuthenticationSimple()) == "secret" {
		w.Write(res)
		return
	}

	log.Printf("Bind failed User=%s, Pass=%s", string(r.Name()), string(r.AuthenticationSimple()))
	res.SetResultCode(ldapserver.LDAPResultInvalidCredentials)
	res.SetDiagnosticMessage("invalid credentials")
	w.Write(res)
}

func handleUserSearch(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	r := m.GetSearchRequest()
	log.Println("User search filter", r.FilterString())

	// one entry
	if r.FilterString() == "(&(objectClass=person)(uid=nahid))" {
		e := ldapserver.NewSearchResultEntry("uid=nahid,ou=users,o=Company")
		e.AddAttribute("cn", "nahid")

		w.Write(e)
	}

	// one entry
	if r.FilterString() == "(&(objectClass=person)(uid=shuvo))" {
		e := ldapserver.NewSearchResultEntry("uid=shuvo,ou=users,o=Company")
		e.AddAttribute("cn", "shuvo")

		w.Write(e)
	}

	// mutliple entry
	if r.FilterString() == "(&(objectClass=person)(id=nahid))" {
		e := ldapserver.NewSearchResultEntry("uid=nahid,ou=users,o=Company")
		e.AddAttribute("cn", "nahid")
		e.AddAttribute("id", "1204")

		e1 := ldapserver.NewSearchResultEntry("uid=shuvo,ou=users,o=Company")
		e1.AddAttribute("cn", "shuvo")
		e1.AddAttribute("id", "1204")

		w.Write(e)
		w.Write(e1)
	}

	res := ldapserver.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
	w.Write(res)
}

func handleGroupSearch(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	r := m.GetSearchRequest()
	log.Println("Group search filter", r.FilterString())

	// one entry
	if r.FilterString() == "(&(objectClass=groupOfNames)(member=uid=nahid,ou=users,o=Company))" {
		e := ldapserver.NewSearchResultEntry("id=1,ou=groups,o=Company")
		e.AddAttribute("cn", "group1")

		e1 := ldapserver.NewSearchResultEntry("id=1,ou=groups,o=Company")
		e1.AddAttribute("cn", "group2")

		w.Write(e)
		w.Write(e1)
	}

	res := ldapserver.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
	w.Write(res)
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
	time.Sleep(2 * time.Second)
	defer srv.stop()

	if secureConn {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(srv.certStore.CACert())
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

	for _, test := range dataset {
		t.Run(serverType+": "+test.testName, func(t *testing.T) {
			t.Log(test)

			serv := s
			serv.opts.UserAttribute = test.userAttribute

			resp, err := serv.Check(base64.StdEncoding.EncodeToString([]byte(test.token)))
			if test.authenticated {
				if assert.Nil(t, err) {
					if resp.Username != test.username {
						t.Errorf("Expected username %v, got %v", test.username, resp.Username)
					}
					if len(resp.Groups) != len(test.groups) {
						t.Errorf("Expected group size %v, got %v", len(test.groups), len(resp.Groups))
					} else {
						if len(resp.Groups) > 0 {
							if !reflect.DeepEqual(resp.Groups, test.groups) {
								t.Errorf("Expected groups %v, got %v", test.groups, resp.Groups)
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
