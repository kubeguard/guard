package lib

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	"github.com/go-ldap/ldap"
	"github.com/spf13/pflag"
	auth "k8s.io/api/authentication/v1beta1"
)

const (
	DefaultUserSearchFilter     string = "(objectClass=person)"
	DefaultGroupSearchFilter    string = "(objectClass=groupOfNames)"
	DefaultUserAttribute        string = "uid"
	DefaultGroupMemberAttribute string = "member"
	DefaultGroupNameAttribute   string = "cn"
)

type LDAPOptions struct {
	ServerAddress        string
	ServerPort           string
	BindDN               string // The connector uses this DN in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.
	BindPassword         string // The connector uses this Password in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.
	UserSearchDN         string // BaseDN to start the search user
	UserSearchFilter     string // filter to apply when searching user, default : (objectClass=person)
	UserAttribute        string // Ldap username attribute, default : uid
	GroupSearchDN        string // BaseDN to start the search group
	GroupSearchFilter    string // filter to apply when searching the groups that user is member of, default : (objectClass=groupOfNames)
	GroupMemberAttribute string // Ldap group member attribute, default: member
	GroupNameAttribute   string // Ldap group name attribute, default: cn
	SkipTLSVerification  bool
	IsSecureLDAP         bool // for LDAP over SSL
	StartTLS             bool // for start tls connection
}

func (s *LDAPOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.ServerAddress, "ldap.server-address", s.ServerAddress, "Host or IP of the LDAP server")
	fs.StringVar(&s.ServerPort, "ldap.server-port", "389", "LDAP server port")
	fs.StringVar(&s.BindDN, "ldap.bind-dn", s.BindDN, "The connector uses this DN in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.")
	fs.StringVar(&s.BindPassword, "ldap.bind-password", s.BindPassword, "The connector uses this password in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.")
	fs.StringVar(&s.UserSearchDN, "ldap.user-search-dn", s.UserSearchDN, "BaseDN to start the search user")
	fs.StringVar(&s.UserSearchFilter, "ldap.user-search-filter", DefaultUserSearchFilter, "Filter to apply when searching user")
	fs.StringVar(&s.UserAttribute, "ldap.user-attribute", DefaultUserAttribute, "Ldap username attribute")
	fs.StringVar(&s.GroupSearchDN, "ldap.group-search-dn", s.GroupSearchDN, "BaseDN to start the search group")
	fs.StringVar(&s.GroupSearchFilter, "ldap.group-search-filter", DefaultGroupSearchFilter, "Filter to apply when searching the groups that user is member of")
	fs.StringVar(&s.GroupMemberAttribute, "ldap.group-member-attribute", DefaultGroupMemberAttribute, "Ldap group member attribute")
	fs.StringVar(&s.GroupNameAttribute, "ldap.group-name-attribute", DefaultGroupNameAttribute, "Ldap group name attribute")
	fs.BoolVar(&s.SkipTLSVerification, "ldap.skip-tls-verification", false, "Skip LDAP server TLS verification, default : false")
	fs.BoolVar(&s.IsSecureLDAP, "ldap.is-secure-ldap", false, "Secure LDAP (LDAPS)")
	fs.BoolVar(&s.StartTLS, "ldap.start-tls", false, "Start tls connection")
}

func (s Server) checkLDAP(token string) (auth.TokenReview, int) {
	username, password, ok := parseEncodedToken(token)
	if !ok {
		return Error("Invalid basic auth token"), http.StatusUnauthorized
	}

	data := auth.TokenReview{}
	tlsConfig := &tls.Config{
		ServerName:         s.LDAP.ServerAddress,
		InsecureSkipVerify: s.LDAP.SkipTLSVerification,
	}
	var (
		err  error
		conn *ldap.Conn
	)
	if s.LDAP.IsSecureLDAP {
		conn, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%s", s.LDAP.ServerAddress, s.LDAP.ServerPort), tlsConfig)
	} else {
		conn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%s", s.LDAP.ServerAddress, s.LDAP.ServerPort))
	}
	if err != nil {
		return Error(fmt.Sprintf("Unable to create ldap connector for %s:%s", s.LDAP.ServerAddress, s.LDAP.ServerPort)), http.StatusInternalServerError
	}
	defer conn.Close()

	if s.LDAP.StartTLS {
		err = conn.StartTLS(tlsConfig)
		if err != nil {
			return Error("Unable to setup TLS connection"), http.StatusInternalServerError
		}
	}

	if s.LDAP.BindDN != "" && s.LDAP.BindPassword != "" {
		err = conn.Bind(s.LDAP.BindDN, s.LDAP.BindPassword)
		if err != nil {
			return Error(err.Error()), http.StatusUnauthorized
		}
	}

	req := s.LDAP.newUserSearchRequest(username)
	res, err := conn.Search(req)
	if err != nil {
		return Error(fmt.Sprintf("Error searching for user %s. Reason: %v", username, err)), http.StatusUnauthorized
	}

	if len(res.Entries) == 0 {
		return Error(fmt.Sprintf("No result for the user search filter '%s'", req.Filter)), http.StatusUnauthorized
	} else if len(res.Entries) > 1 {
		log.Infof(fmt.Sprintf("Multiple entries found for the user search filter '%s': %+v", req.Filter, res.Entries))
		return Error(fmt.Sprintf("Multiple entries found for the user search filter '%s'", req.Filter)), http.StatusUnauthorized
	}

	userDN := res.Entries[0].DN
	// authenticate user
	err = conn.Bind(userDN, password)
	if err != nil {
		return Error(err.Error()), http.StatusUnauthorized
	}

	//rebind
	if s.LDAP.BindDN != "" && s.LDAP.BindPassword != "" {
		err = conn.Bind(s.LDAP.BindDN, s.LDAP.BindPassword)
		if err != nil {
			return Error(err.Error()), http.StatusUnauthorized
		}
	}

	// user group list
	req = s.LDAP.newGroupSearchRequest(userDN)
	res, err = conn.Search(req)
	if err != nil {
		return Error(fmt.Sprintf("Error searching for user's group for %s : %v", userDN, err)), http.StatusUnauthorized
	}
	var groups []string
	//default use `cn` as group name
	for _, en := range res.Entries {
		for _, g := range en.Attributes {
			if g.Name == s.LDAP.GroupNameAttribute {
				if len(g.Values) == 0 {
					return Error(fmt.Sprintf("cn not provided for %s", en.DN)), http.StatusUnauthorized
				} else {
					groups = append(groups, g.Values[0])
				}
			}
		}
	}

	data.Status.Authenticated = true
	data.Status.User.Username = username
	data.Status.User.Groups = groups
	return data, http.StatusOK
}

// request to search user
func (s *LDAPOptions) newUserSearchRequest(username string) *ldap.SearchRequest {
	userFilter := fmt.Sprintf("(&%s(%s=%s))", s.UserSearchFilter, s.UserAttribute, username)
	return &ldap.SearchRequest{
		BaseDN:       s.UserSearchDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    2, //limit number of entries in result
		TimeLimit:    10,
		TypesOnly:    false,
		Filter:       userFilter, //filter default format : (&(objectClass=person)(uid=%s))
	}
}

// request to get user group list
func (s *LDAPOptions) newGroupSearchRequest(userDN string) *ldap.SearchRequest {
	groupFilter := fmt.Sprintf("(&%s(%s=%s))", s.GroupSearchFilter, s.GroupMemberAttribute, userDN)
	return &ldap.SearchRequest{
		BaseDN:       s.GroupSearchDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0, //limit number of entries in result, 0 values means no limitations
		TimeLimit:    10,
		TypesOnly:    false,
		Filter:       groupFilter, //filter default format : (&(objectClass=groupOfNames)(member=%s))
		Attributes:   []string{s.GroupNameAttribute},
	}
}

// parseEncodedToken parses base64 encode token
// "dXNlcjE6MTIzNA==" returns ("user1", "1234", true).
func parseEncodedToken(token string) (username, password string, ok bool) {
	c, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}
