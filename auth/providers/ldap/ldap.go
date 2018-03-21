package ldap

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/appscode/guard/auth"
	"github.com/go-ldap/ldap"
	"github.com/pkg/errors"
	"gopkg.in/jcmturner/gokrb5.v4/messages"
	"gopkg.in/jcmturner/gokrb5.v4/service"
	authv1 "k8s.io/api/authentication/v1"
)

const (
	OrgType = "ldap"

	DefaultUserSearchFilter     = "(objectClass=person)"
	DefaultGroupSearchFilter    = "(objectClass=groupOfNames)"
	DefaultUserAttribute        = "uid"
	DefaultGroupMemberAttribute = "member"
	DefaultGroupNameAttribute   = "cn"

	AuthChoiceSimpleAuthentication = 0
	AuthChoiceKerberos             = 1
)

func init() {
	auth.SupportedOrgs = append(auth.SupportedOrgs, OrgType)
}

type Authenticator struct {
	opts  Options
	token string
}

func New(opts Options, token string) auth.Interface {
	return &Authenticator{
		opts:  opts,
		token: token,
	}
}

func (g Authenticator) UID() string {
	return OrgType
}

func (s Authenticator) Check() (*authv1.UserInfo, error) {
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

	username, err := s.authenticateUser(conn, s.token)
	if err != nil {
		return nil, errors.Wrap(err, "authentication failed")
	}

	if s.opts.AuthenticationChoice == AuthChoiceSimpleAuthentication {
		// rebind, as in simple authentication we bind using username, password
		if s.opts.BindDN != "" && s.opts.BindPassword != "" {
			err = conn.Bind(s.opts.BindDN, s.opts.BindPassword)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	userDN, err := s.getUserDN(conn, username)
	if err != nil {
		return nil, errors.Wrap(err, "error when getting user DN")
	}

	// user group list
	req := s.opts.newGroupSearchRequest(userDN)
	res, err := conn.Search(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error searching for user's group for %s", userDN)
	}

	var groups []string
	//default use `cn` as group name
	for _, en := range res.Entries {
		for _, g := range en.Attributes {
			if g.Name == s.opts.GroupNameAttribute {
				if len(g.Values) == 0 {
					return nil, errors.Errorf("%s not provided for %s", s.opts.GroupNameAttribute, en.DN)
				} else {
					groups = append(groups, g.Values[0])
				}
			}
		}
	}

	resp := &authv1.UserInfo{}
	resp.Username = username
	resp.Groups = groups
	return resp, nil
}

func (s Authenticator) authenticateUser(conn *ldap.Conn, token string) (string, error) {
	if s.opts.AuthenticationChoice == AuthChoiceSimpleAuthentication {
		//simple authentication
		username, password, ok := parseEncodedToken(token)
		if !ok {
			return "", errors.New("Invalid basic auth token")
		}

		userDN, err := s.getUserDN(conn, username)
		if err != nil {
			return "", errors.WithStack(err)
		}

		// authenticate user
		err = conn.Bind(userDN, password)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return username, nil

	} else if s.opts.AuthenticationChoice == AuthChoiceKerberos {
		// kerberos
		data, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			return "", errors.Wrap(err, "unable to decode token")
		}

		apReq := &messages.APReq{}
		err = apReq.Unmarshal(data)
		if err != nil {
			return "", errors.Wrap(err, "unable to unmarshall")
		}

		if ok, creds, err := service.ValidateAPREQ(*apReq, s.opts.keytab, s.opts.ServiceAccountName, "", false); ok {
			return creds.Username, nil
		} else {
			return "", err
		}

	} else {
		return "", errors.New("authentication choice invalid")
	}
}

func (s Authenticator) getUserDN(conn *ldap.Conn, username string) (string, error) {
	req := s.opts.newUserSearchRequest(username)

	res, err := conn.Search(req)
	if err != nil {
		return "", errors.Wrapf(err, "error searching for user %s", username)
	}

	if len(res.Entries) == 0 {
		return "", errors.Errorf("No result for the user search filter '%s'", req.Filter)
	} else if len(res.Entries) > 1 {
		return "", errors.Errorf("Multiple entries found for the user search filter '%s'", req.Filter)
	}

	userDN := res.Entries[0].DN

	return userDN, nil
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
