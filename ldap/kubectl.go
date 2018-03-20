package ldap

import (
	"encoding/base64"

	"github.com/pkg/errors"
	"gopkg.in/jcmturner/gokrb5.v4/client"
	"gopkg.in/jcmturner/gokrb5.v4/crypto"
	"gopkg.in/jcmturner/gokrb5.v4/messages"
	"gopkg.in/jcmturner/gokrb5.v4/types"
)

func GetKerberosToken(userName, userPwd, realm, krb5Config, ServicePrincipalName string) (string, error) {
	cl := client.NewClientWithPassword(userName, realm, userPwd)

	c, err := cl.LoadConfig(krb5Config)
	if err != nil {
		return "", errors.Wrap(err, "failed to load krb5 config file")
	}

	// Active Directory does not commonly support FAST negotiation so you will need to disable this on the client.
	// If this is the case you will see this error: KDC did not respond appropriately  to FAST negotiation To resolve
	// this disable PA-FX-Fast on the client before performing Login()
	c.GoKrb5Conf.DisablePAFXFast = true

	err = c.Login()
	if err != nil {
		return "", errors.Wrap(err, "login unsuccessful")
	}

	tkt, key, err := c.GetServiceTicket(ServicePrincipalName)
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
