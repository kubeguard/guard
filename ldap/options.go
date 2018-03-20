package ldap

import (
	"crypto/x509"
	"fmt"

	"github.com/go-ldap/ldap"
	"github.com/spf13/pflag"
)

type Options struct {
	ServerAddress string

	ServerPort string

	// The connector uses this DN in credentials to search for users and groups.
	// Not required if the LDAP server provides access for anonymous auth.
	BindDN string

	// The connector uses this Password in credentials to search for users and groups.
	// Not required if the LDAP server provides access for anonymous auth.
	BindPassword string

	// BaseDN to start the search user
	UserSearchDN string

	// filter to apply when searching user
	// default : (objectClass=person)
	UserSearchFilter string

	// Ldap username attribute
	// default : uid
	UserAttribute string

	//BaseDN to start the search group
	GroupSearchDN string

	// filter to apply when searching the groups that user is member of
	// default : (objectClass=groupOfNames)
	GroupSearchFilter string

	// Ldap group member attribute
	// default: member
	GroupMemberAttribute string

	// Ldap group name attribute
	// default: cn
	GroupNameAttribute string

	SkipTLSVerification bool

	// for LDAP over SSL
	IsSecureLDAP bool

	// for start tls connection
	StartTLS bool

	// path to the caCert file, needed for self signed server certificate
	CaCertFile string

	CaCertPool *x509.CertPool

	// LDAP user authentication mechanism
	// 0 for simple authentication
	// 1 for kerberos(via GSSAPI)
	AuthenticationChoice int

	// path to the keytab file
	// it's contain LDAP service principal keys
	// required for kerberos
	// default : 0
	KeytabFile string

	// The serviceAccountName needs to be defined when using Active Directory
	// where the SPN is mapped to a user account. If this is not required it
	// should be set to an empty string ""
	// default : ""
	ServiceAccountName string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ServerAddress, "ldap.server-address", o.ServerAddress, "Host or IP of the LDAP server")
	fs.StringVar(&o.ServerPort, "ldap.server-port", "389", "LDAP server port")
	fs.StringVar(&o.BindDN, "ldap.bind-dn", o.BindDN, "The connector uses this DN in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.")
	fs.StringVar(&o.BindPassword, "ldap.bind-password", o.BindPassword, "The connector uses this password in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.")
	fs.StringVar(&o.UserSearchDN, "ldap.user-search-dn", o.UserSearchDN, "BaseDN to start the search user")
	fs.StringVar(&o.UserSearchFilter, "ldap.user-search-filter", DefaultUserSearchFilter, "Filter to apply when searching user")
	fs.StringVar(&o.UserAttribute, "ldap.user-attribute", DefaultUserAttribute, "Ldap username attribute")
	fs.StringVar(&o.GroupSearchDN, "ldap.group-search-dn", o.GroupSearchDN, "BaseDN to start the search group")
	fs.StringVar(&o.GroupSearchFilter, "ldap.group-search-filter", DefaultGroupSearchFilter, "Filter to apply when searching the groups that user is member of")
	fs.StringVar(&o.GroupMemberAttribute, "ldap.group-member-attribute", DefaultGroupMemberAttribute, "Ldap group member attribute")
	fs.StringVar(&o.GroupNameAttribute, "ldap.group-name-attribute", DefaultGroupNameAttribute, "Ldap group name attribute")
	fs.BoolVar(&o.SkipTLSVerification, "ldap.skip-tls-verification", false, "Skip LDAP server TLS verification, default : false")
	fs.BoolVar(&o.IsSecureLDAP, "ldap.is-secure-ldap", false, "Secure LDAP (LDAPS)")
	fs.BoolVar(&o.StartTLS, "ldap.start-tls", false, "Start tls connection")
	fs.StringVar(&o.CaCertFile, "ldap.ca-cert-file", "", "ca cert file that used for self signed server certificate")
	fs.IntVar(&o.AuthenticationChoice, "ldap.auth-choice", 0, "LDAP user authentication mechanism, 0 for simple authentication, 1 for kerberos(via GSSAPI)")
	fs.StringVar(&o.KeytabFile, "ldap.keytab-file", "", "path to the keytab file, it's contain LDAP service principal keys")
	fs.StringVar(&o.ServiceAccountName, "ldap.service-account", "", "service account name")
}

func (o Options) ToArgs() []string {
	var args []string
	if o.ServerAddress != "" {
		args = append(args, fmt.Sprintf("--ldap.server-address=%s", o.ServerAddress))
	}
	if o.ServerPort != "" {
		args = append(args, fmt.Sprintf("--ldap.server-port=%s", o.ServerPort))
	}
	if o.BindDN != "" {
		args = append(args, fmt.Sprintf("--ldap.bind-dn=%s", o.BindDN))
	}
	if o.BindPassword != "" {
		args = append(args, fmt.Sprintf("--ldap.bind-password=%s", o.BindPassword))
	}
	if o.UserSearchDN != "" {
		args = append(args, fmt.Sprintf("--ldap.user-search-dn=%s", o.UserSearchDN))
	}
	if o.UserSearchFilter != "" {
		args = append(args, fmt.Sprintf("--ldap.user-search-filter=%s", o.UserSearchFilter))
	}
	if o.UserSearchFilter != "" {
		args = append(args, fmt.Sprintf("--ldap.user-attribute=%s", o.UserAttribute))
	}
	if o.GroupSearchDN != "" {
		args = append(args, fmt.Sprintf("--ldap.group-search-dn=%s", o.GroupSearchDN))
	}
	if o.GroupSearchFilter != "" {
		args = append(args, fmt.Sprintf("--ldap.group-search-filter=%s", o.GroupSearchFilter))
	}
	if o.GroupMemberAttribute != "" {
		args = append(args, fmt.Sprintf("--ldap.group-member-attribute=%s", o.GroupMemberAttribute))
	}
	if o.GroupNameAttribute != "" {
		args = append(args, fmt.Sprintf("--ldap.group-name-attribute=%s", o.GroupNameAttribute))
	}
	if o.SkipTLSVerification {
		args = append(args, "--ldap.skip-tls-verification")
	}
	if o.IsSecureLDAP {
		args = append(args, "--ldap.is-secure-ldap")
	}
	if o.StartTLS {
		args = append(args, "--ldap.start-tls")
	}
	if o.CaCertFile != "" {
		args = append(args, fmt.Sprintf("--ldap.ca-cert-file=/etc/guard/ldap/ca.crt"))
	}
	if o.ServiceAccountName != "" {
		args = append(args, fmt.Sprintf("--ldap.service-account=%s", o.ServiceAccountName))
	}
	if o.KeytabFile != "" {
		args = append(args, fmt.Sprintf("--ldap.keytab-file=/etc/guard/ldap/krb5.keytab"))
	}
	args = append(args, fmt.Sprintf("--ldap.auth-choice=%v", o.AuthenticationChoice))

	return args
}

// request to search user
func (o *Options) newUserSearchRequest(username string) *ldap.SearchRequest {
	userFilter := fmt.Sprintf("(&%s(%s=%s))", o.UserSearchFilter, o.UserAttribute, username)
	return &ldap.SearchRequest{
		BaseDN:       o.UserSearchDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    2, // limit number of entries in result
		TimeLimit:    10,
		TypesOnly:    false,
		Filter:       userFilter, // filter default format : (&(objectClass=person)(uid=%s))
	}
}

// request to get user group list
func (o *Options) newGroupSearchRequest(userDN string) *ldap.SearchRequest {
	groupFilter := fmt.Sprintf("(&%s(%s=%s))", o.GroupSearchFilter, o.GroupMemberAttribute, userDN)
	return &ldap.SearchRequest{
		BaseDN:       o.GroupSearchDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0, // limit number of entries in result, 0 values means no limitations
		TimeLimit:    10,
		TypesOnly:    false,
		Filter:       groupFilter, // filter default format : (&(objectClass=groupOfNames)(member=%s))
		Attributes:   []string{o.GroupNameAttribute},
	}
}

func (o *Options) Validate() []error {
	return nil
}
