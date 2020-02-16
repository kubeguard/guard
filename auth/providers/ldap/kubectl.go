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
	"encoding/base64"

	"github.com/appscode/guard/util/kubeconfig"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"gopkg.in/jcmturner/gokrb5.v4/client"
	"gopkg.in/jcmturner/gokrb5.v4/crypto"
	"gopkg.in/jcmturner/gokrb5.v4/messages"
	"gopkg.in/jcmturner/gokrb5.v4/types"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type TokenOptions struct {
	Username string

	UserPassword string

	// set the realm to empty string to use the default realm from config
	Realm string

	Krb5configFile string

	ServicePrincipalName string

	// Active Directory does not commonly support FAST negotiation so you will need to disable this on the client.
	// If this is the case you will see this error: KDC did not respond appropriately  to FAST negotiation To resolve
	// this disable PA-FX-Fast on the client before performing Login()
	DisablePAFXFast bool

	// LDAP user authentication mechanism
	// 0 for simple authentication
	// 1 for kerberos(via GSSAPI)
	// default: 0 (simple authentication)
	AuthenticationChoice int
}

func (t *TokenOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&t.Username, "ldap.username", t.Username, "Username")
	fs.StringVar(&t.UserPassword, "ldap.password", t.UserPassword, "Password")
	fs.StringVar(&t.Realm, "ldap.realm", t.Realm, "Realm, set the realm to empty string to use the default realm from config")
	fs.StringVar(&t.Krb5configFile, "ldap.krb5-config", "/etc/krb5.conf", "Path to the kerberos configuration file")
	fs.StringVar(&t.ServicePrincipalName, "ldap.spn", t.ServicePrincipalName, "Service principal name")
	fs.BoolVar(&t.DisablePAFXFast, "ldap.disable-pa-fx-fast", true, "Disable PA-FX-Fast, Active Directory does not commonly support FAST negotiation so you will need to disable this on the client")
	fs.IntVar(&t.AuthenticationChoice, "ldap.auth-choice", 0, "LDAP user authentication mechanism, 0 for simple authentication, 1 for kerberos(via GSSAPI)")
}

func (t *TokenOptions) IssueToken() error {
	var (
		token string
		err   error
	)

	err = t.Validate()
	if err != nil {
		return err
	}

	switch t.AuthenticationChoice {
	case 0:
		token = t.getSimpleAuthToken()
	case 1:
		// ref: https://www.youtube.com/watch?v=KD2Q-2ToloE
		token, err = t.getKerberosToken()
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid authentication choice")
	}

	return t.addAuthInfo(token)
}

func (t *TokenOptions) Validate() error {
	if t.Username == "" {
		return errors.New("username is required")
	}
	if t.UserPassword == "" {
		return errors.New("password is required")
	}
	if t.AuthenticationChoice == 1 && t.ServicePrincipalName == "" {
		return errors.New("service principal is required")
	}
	return nil
}

func (t *TokenOptions) getSimpleAuthToken() string {
	return base64.StdEncoding.EncodeToString([]byte(t.Username + ":" + t.UserPassword))
}

func (t *TokenOptions) getKerberosToken() (string, error) {
	cl := client.NewClientWithPassword(t.Username, t.Realm, t.UserPassword)

	c, err := cl.LoadConfig(t.Krb5configFile)
	if err != nil {
		return "", errors.Wrap(err, "failed to load krb5 config file")
	}

	c.GoKrb5Conf.DisablePAFXFast = t.DisablePAFXFast

	err = c.Login()
	if err != nil {
		return "", errors.Wrap(err, "login unsuccessful")
	}

	tkt, key, err := c.GetServiceTicket(t.ServicePrincipalName)
	if err != nil {
		return "", errors.Wrap(err, "failed to get service ticket")
	}

	auth, err := types.NewAuthenticator(c.Credentials.Realm, c.Credentials.CName)
	if err != nil {
		return "", errors.Wrap(err, "failed to create authenticator")
	}

	etype, err := crypto.GetEtype(key.KeyType)
	if err != nil {
		return "", errors.Wrap(err, "failed to get encryption type")
	}

	err = auth.GenerateSeqNumberAndSubKey(key.KeyType, etype.GetKeyByteSize())
	if err != nil {
		return "", errors.Wrap(err, "failed to generate sequence number and subkey")
	}

	apReq, err := messages.NewAPReq(tkt, key, auth)
	if err != nil {
		return "", errors.Wrap(err, "failed to create AP_REQ message")
	}

	data, err := apReq.Marshal()
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal AP_REQ message")
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (t *TokenOptions) addAuthInfo(token string) error {
	authInfo := &clientcmdapi.AuthInfo{
		Token: token,
	}

	return kubeconfig.AddAuthInfo(t.Username, authInfo)
}
