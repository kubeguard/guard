package server

import (
	"fmt"

	"github.com/spf13/pflag"
)

const (
	ServingPort = 8443
)

type SecureServingOptions struct {
	SecureAddr string
	CACertFile string
	CertFile   string
	KeyFile    string
}

func NewSecureServingOptions() SecureServingOptions {
	return SecureServingOptions{
		SecureAddr: fmt.Sprintf(":%d", ServingPort),
	}
}

func (o *SecureServingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.SecureAddr, "secure-addr", o.SecureAddr, "host:port used to serve secure apis")

	fs.StringVar(&o.CACertFile, "tls-ca-file", o.CACertFile, "File containing CA certificate")
	fs.StringVar(&o.CertFile, "tls-cert-file", o.CertFile, "File container server TLS certificate")
	fs.StringVar(&o.KeyFile, "tls-private-key-file", o.KeyFile, "File containing server TLS private key")
}

func (o SecureServingOptions) ToArgs() []string {
	var args []string

	args = append(args, "--tls-ca-file=/etc/guard/pki/ca.crt")
	args = append(args, "--tls-cert-file=/etc/guard/pki/tls.crt")
	args = append(args, "--tls-private-key-file=/etc/guard/pki/tls.key")

	return args
}

func (o *SecureServingOptions) Validate() []error {
	return nil
}

func (o SecureServingOptions) UseTLS() bool {
	return o.CACertFile != "" && o.CertFile != "" && o.KeyFile != ""
}
