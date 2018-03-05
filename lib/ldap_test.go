package lib

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/appscode/kutil/tools/certstore"
	"github.com/go-ldap/ldap"
	"github.com/spf13/afero"
	"github.com/vjeantet/ldapserver"
	"k8s.io/client-go/util/cert"
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
				err = s.server.ListenAndServe("127.0.0.1:8089", func(s *ldapserver.Server) {
					s.Listener = tls.NewListener(s.Listener, tlsConfig)
				})
			}
		} else {
			err = s.server.ListenAndServe("127.0.0.1:8089")
		}
		log.Println("LDAP Server: ", err)
	}()

	<-s.stopCh
	close(s.stopCh)
	s.server.Stop()
}

func (s *ldapServer) stop() {
	s.stopCh <- true
}

// getTLSconfig returns a tls configuration used
// to build a TLSlistener for TLS or StartTLS
func (s *ldapServer) getTLSconfig() (*tls.Config, error) {
	srvCert, srvKey, err := s.certStore.NewServerCertPair("server", cert.AltNames{IPs: []net.IP{net.ParseIP("127.0.0.1")}})
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
		ServerName:   "127.0.0.1",
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
		store, err := certstore.NewCertStore(afero.NewMemMapFs(), filepath.Join("", "certs"), "test")
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

		e1 := ldapserver.NewSearchResultEntry("id=1,ou=groupss,o=Company")
		e1.AddAttribute("cn", "group2")

		w.Write(e)
		w.Write(e1)
	}

	res := ldapserver.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
	w.Write(res)
}

func TestCheckLdapInSecure(t *testing.T) {
	srv, err := ldapServerSetup(false, "o=Company,ou=users", "o=Company,ou=groups")
	if err != nil {
		t.Fatal(err)
	}

	go srv.start()
	time.Sleep(1 * time.Second)
	defer srv.stop()

	opts := LDAPOptions{
		ServerAddress:        "127.0.0.1",
		ServerPort:           "8089",
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
	s := Server{
		LDAP: opts,
	}

	// authenticated : true
	t.Run("scenario 1", func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid:secret")))
		if status != http.StatusOK {
			t.Errorf("Expected authentication true, got false. reason: %v", resp.Status.Error)
		}

		u := resp.Status.User
		if u.Username != "nahid" {
			t.Errorf("Expected username %v, got %v", "nahid", u.Username)
		}
		if g := []string{"group1", "group2"}; !reflect.DeepEqual(u.Groups, g) {
			t.Errorf("Expected groups %v, got %v", g, u.Groups)
		}
	})

	// error expected
	// multiple entry when searching userDN
	t.Run("scenario 2", func(t *testing.T) {
		serv := s
		serv.LDAP.UserAttribute = "id"
		resp, status := serv.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid:secret")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
		// restoring to previous value
		//s.LDAP.UserAttribute = DefaultUserAttribute
	})

	// error expected
	// empty entry when searching userDN
	t.Run("scenario 3", func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid1:secret")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// error expected
	// invalid token
	t.Run("scenario 4", func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("invalid_token")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// error expected
	// invalid token
	t.Run("scenario 5", func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("invalid_token")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// error expected
	// failed to authenticate user
	t.Run("scenario 6", func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid:12345")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// empty groups
	t.Run("scenario 7", func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("shuvo:secret")))
		if status != http.StatusOK {
			t.Errorf("Expected authentication true, got false. reason: %v", resp.Status.Error)
		}

		u := resp.Status.User
		if u.Username != "shuvo" {
			t.Errorf("Expected username %v, got %v", "nahid", u.Username)
		}
		if len(u.Groups) > 0 {
			t.Errorf("Expected empty groups, got %v", u.Groups)
		}
	})

}

func TestCheckLdapSecure(t *testing.T) {
	srv, err := ldapServerSetup(true, "o=Company,ou=users", "o=Company,ou=groups")
	if err != nil {
		t.Fatal(err)
	}

	go srv.start()
	time.Sleep(1 * time.Second)
	defer srv.stop()

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(srv.certStore.CACert())

	opts := LDAPOptions{
		ServerAddress:        "127.0.0.1",
		ServerPort:           "8089",
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
		caCertPool:           caCertPool,
		CaCertFile:           "/test/certs/ca.file",
	}
	s := Server{
		LDAP: opts,
	}

	testType := "Secure LDAP"

	// authenticated : true
	t.Run(fmt.Sprintf("%v : scenario 1", testType), func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid:secret")))
		if status != http.StatusOK {
			t.Errorf("Expected authentication true, got false. reason: %v", resp.Status.Error)
		}

		u := resp.Status.User
		if u.Username != "nahid" {
			t.Errorf("Expected username %v, got %v", "nahid", u.Username)
		}
		if g := []string{"group1", "group2"}; !reflect.DeepEqual(u.Groups, g) {
			t.Errorf("Expected groups %v, got %v", g, u.Groups)
		}
	})

	// error expected
	// multiple entry when searching userDN
	t.Run(fmt.Sprintf("%v : scenario 2", testType), func(t *testing.T) {
		serv := s
		serv.LDAP.UserAttribute = "id"
		resp, status := serv.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid:secret")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// error expected
	// empty entry when searching userDN
	t.Run(fmt.Sprintf("%v : scenario 3", testType), func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid1:secret")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// error expected
	// invalid token
	t.Run(fmt.Sprintf("%v : scenario 4", testType), func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("invalid_token")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// error expected
	// invalid token
	t.Run(fmt.Sprintf("%v : scenario 5", testType), func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("invalid_token")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// error expected
	// failed to authenticate user
	t.Run(fmt.Sprintf("%v : scenario 6", testType), func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("nahid:12345")))
		if status != http.StatusUnauthorized {
			t.Errorf("Expected authentication false, got true")
		}

		if resp.Status.Error == "" {
			t.Errorf("Expected non empty error message")
		}
	})

	// empty groups
	t.Run(fmt.Sprintf("%v : scenario 7", testType), func(t *testing.T) {
		resp, status := s.checkLDAP(base64.StdEncoding.EncodeToString([]byte("shuvo:secret")))
		if status != http.StatusOK {
			t.Errorf("Expected authentication true, got false. reason: %v", resp.Status.Error)
		}

		u := resp.Status.User
		if u.Username != "shuvo" {
			t.Errorf("Expected username %v, got %v", "shuvou", u.Username)
		}
		if len(u.Groups) > 0 {
			t.Errorf("Expected empty groups, got %v", u.Groups)
		}
	})

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
