package ldap

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/appscode/go/types"
	"github.com/go-ldap/ldap"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"gopkg.in/jcmturner/gokrb5.v4/keytab"
	"k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	AuthenticationChoice AuthChoice

	// path to the keytab file
	// it's contain LDAP service principal keys
	// required for kerberos
	// default : 0
	KeytabFile string

	// keytab contains service principal and encryption key
	keytab keytab.Keytab

	// The serviceAccountName needs to be defined when using Active Directory
	// where the SPN is mapped to a user account. If this is not required it
	// should be set to an empty string ""
	// default : ""
	ServiceAccountName string
}

func NewOptions() Options {
	return Options{
		BindDN:       os.Getenv("LDAP_BIND_DN"),
		BindPassword: os.Getenv("LDAP_BIND_PASSWORD"),
	}
}

// if ca cert is provided then create CA Cert Pool
// if keytab file is provides then load it
func (o *Options) Configure() error {
	// caCertPool for self signed LDAP sever certificate
	if o.CaCertFile != "" {
		caCert, err := ioutil.ReadFile(o.CaCertFile)
		if err != nil {
			return errors.Wrap(err, "unable to read ca cert file")
		}
		o.CaCertPool = x509.NewCertPool()
		o.CaCertPool.AppendCertsFromPEM(caCert)
		ok := o.CaCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return errors.New("Failed to add CA cert in CertPool for LDAP")
		}
	}

	// keytab required for kerberos
	if o.AuthenticationChoice == AuthChoiceKerberos {
		var err error
		if o.KeytabFile != "" {
			return errors.New("keytab not provided")
		}

		o.keytab, err = keytab.Load(o.KeytabFile)
		if err != nil {
			return errors.Wrap(err, "unable to parse keytab file")
		}
	}
	return nil
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
	fs.Var(&o.AuthenticationChoice, "ldap.auth-choice", "LDAP user authentication mechanisms Simple/Kerberos(via GSSAPI)")
	fs.StringVar(&o.KeytabFile, "ldap.keytab-file", "", "path to the keytab file, it's contain LDAP service principal keys")
	fs.StringVar(&o.ServiceAccountName, "ldap.service-account", "", "service account name")
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
	var errs []error
	if o.ServerAddress == "" {
		errs = append(errs, errors.New("ldap.server-address must be non-empty"))
	}
	if o.ServerPort == "" {
		errs = append(errs, errors.New("ldap.server-port must be non-empty"))
	}
	if o.UserSearchDN == "" {
		errs = append(errs, errors.New("ldap.user-search-dn must be non-empty"))
	}
	if o.UserAttribute == "" {
		errs = append(errs, errors.New("ldap.user-attribute must be non-empty"))
	}
	if o.GroupSearchDN == "" {
		errs = append(errs, errors.New("ldap.group-search-dn must be non-empty"))
	}
	if o.GroupMemberAttribute == "" {
		errs = append(errs, errors.New("ldap.group-member-attribute must be non-empty"))
	}
	if o.GroupNameAttribute == "" {
		errs = append(errs, errors.New("ldap.group-name-attribute must be non-empty"))
	}
	if o.IsSecureLDAP && o.StartTLS {
		errs = append(errs, errors.New("ldap.is-secure-ldap and ldap.start-tls both can not be true at the same time"))
	}
	if o.AuthenticationChoice == AuthChoiceKerberos && o.KeytabFile == "" {
		errs = append(errs, errors.New("for kerberos ldap.keytab-file must be non-empty"))
	}
	return errs
}

func (o Options) Apply(d *v1beta1.Deployment) (extraObjs []runtime.Object, err error) {
	container := d.Spec.Template.Spec.Containers[0]

	// create auth secret
	ldapData := map[string][]byte{
		"bind-dn":       []byte(o.BindDN), // username kept in secret, since password is in secret
		"bind-password": []byte(o.BindPassword),
	}
	if o.CaCertFile != "" {
		cert, err := ioutil.ReadFile(o.CaCertFile)
		if err != nil {
			return nil, err
		}
		ldapData["ca.crt"] = cert
	}
	if o.KeytabFile != "" {
		key, err := ioutil.ReadFile(o.KeytabFile)
		if err != nil {
			return nil, err
		}
		ldapData["krb5.keytab"] = key
	}
	authSecret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-ldap-auth",
			Namespace: d.Namespace,
			Labels:    d.Labels,
		},
		Data: ldapData,
	}
	extraObjs = append(extraObjs, authSecret)

	// mount auth secret into deployment
	volMount := core.VolumeMount{
		Name:      authSecret.Name,
		MountPath: "/etc/guard/auth/ldap",
	}
	container.VolumeMounts = append(container.VolumeMounts, volMount)

	vol := core.Volume{
		Name: authSecret.Name,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName:  authSecret.Name,
				DefaultMode: types.Int32P(0444),
			},
		},
	}
	d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, vol)

	// use auth secret in container[0] args
	container.Env = append(container.Env,
		core.EnvVar{
			Name: "LDAP_BIND_DN",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: authSecret.Name,
					},
					Key: "bind-dn",
				},
			},
		},
		core.EnvVar{
			Name: "LDAP_BIND_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: authSecret.Name,
					},
					Key: "bind-password",
				},
			},
		},
	)

	args := container.Args
	if o.ServerAddress != "" {
		args = append(args, fmt.Sprintf("--ldap.server-address=%s", o.ServerAddress))
	}
	if o.ServerPort != "" {
		args = append(args, fmt.Sprintf("--ldap.server-port=%s", o.ServerPort))
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
		args = append(args, fmt.Sprintf("--ldap.ca-cert-file=/etc/guard/auth/ldap/ca.crt"))
	}
	if o.ServiceAccountName != "" {
		args = append(args, fmt.Sprintf("--ldap.service-account=%s", o.ServiceAccountName))
	}
	if o.KeytabFile != "" {
		args = append(args, fmt.Sprintf("--ldap.keytab-file=/etc/guard/auth/ldap/krb5.keytab"))
	}
	args = append(args, fmt.Sprintf("--ldap.auth-choice=%v", o.AuthenticationChoice))

	container.Args = args
	d.Spec.Template.Spec.Containers[0] = container

	return extraObjs, nil
}
