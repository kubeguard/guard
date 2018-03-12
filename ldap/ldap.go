package ldap

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap"
	"github.com/pkg/errors"
	auth "k8s.io/api/authentication/v1"
)

const (
	OrgType = "ldap"

	DefaultUserSearchFilter     = "(objectClass=person)"
	DefaultGroupSearchFilter    = "(objectClass=groupOfNames)"
	DefaultUserAttribute        = "uid"
	DefaultGroupMemberAttribute = "member"
	DefaultGroupNameAttribute   = "cn"
)

type Authenticator struct {
	opts Options
}

func New(opts Options) *Authenticator {
	return &Authenticator{
		opts: opts,
	}
}

func (s Authenticator) Check(token string) (*auth.UserInfo, error) {
	username, password, ok := parseEncodedToken(token)
	if !ok {
		return nil, errors.New("Invalid basic auth token")
	}

	var (
		err  error
		conn *ldap.Conn
	)

	tlsConfig := &tls.Config{
		ServerName:         s.opts.ServerAddress,
		InsecureSkipVerify: s.opts.SkipTLSVerification,
	}

	if s.opts.CaCertFile != "" {
		tlsConfig.RootCAs = s.opts.CaCertPool
	}

	if s.opts.IsSecureLDAP {
		conn, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%s", s.opts.ServerAddress, s.opts.ServerPort), tlsConfig)
	} else {
		conn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%s", s.opts.ServerAddress, s.opts.ServerPort))
	}
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create ldap connector for %s:%s", s.opts.ServerAddress, s.opts.ServerPort)
	}
	defer conn.Close()

	if s.opts.StartTLS {
		err = conn.StartTLS(tlsConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to setup TLS connection")
		}
	}

	if s.opts.BindDN != "" && s.opts.BindPassword != "" {
		err = conn.Bind(s.opts.BindDN, s.opts.BindPassword)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	req := s.opts.newUserSearchRequest(username)
	res, err := conn.Search(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error searching for user %s", username)
	}

	if len(res.Entries) == 0 {
		return nil, errors.Errorf("No result for the user search filter '%s'", req.Filter)
	} else if len(res.Entries) > 1 {
		return nil, errors.Errorf("Multiple entries found for the user search filter '%s'", req.Filter)
	}

	userDN := res.Entries[0].DN
	// authenticate user
	err = conn.Bind(userDN, password)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	//rebind
	if s.opts.BindDN != "" && s.opts.BindPassword != "" {
		err = conn.Bind(s.opts.BindDN, s.opts.BindPassword)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	// user group list
	req = s.opts.newGroupSearchRequest(userDN)
	res, err = conn.Search(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error searching for user's group for %s", userDN)
	}
	var groups []string
	//default use `cn` as group name
	for _, en := range res.Entries {
		for _, g := range en.Attributes {
			if g.Name == s.opts.GroupNameAttribute {
				if len(g.Values) == 0 {
					return nil, errors.Errorf("cn not provided for %s", en.DN)
				} else {
					groups = append(groups, g.Values[0])
				}
			}
		}
	}

	resp := &auth.UserInfo{}
	resp.Username = username
	resp.Groups = groups
	return resp, nil
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
