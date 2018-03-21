package installer

import (
	"github.com/appscode/guard/auth"
	"github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/auth/providers/google"
	"github.com/appscode/guard/auth/providers/ldap"
	"github.com/appscode/guard/auth/providers/token"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Options struct {
	pkiDir          string
	namespace       string
	addr            string
	runOnMaster     bool
	privateRegistry string
	imagePullSecret string

	Token  token.Options
	Google google.Options
	Azure  azure.Options
	LDAP   ldap.Options
}

func New() Options {
	return Options{
		pkiDir:          auth.DefaultPKIDir,
		namespace:       metav1.NamespaceSystem,
		addr:            "10.96.10.96:443",
		privateRegistry: "appscode",
		runOnMaster:     true,
		Token:           token.NewOptions(),
		Google:          google.NewOptions(),
		Azure:           azure.NewOptions(),
		LDAP:            ldap.NewOptions(),
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.pkiDir, "pki-dir", o.pkiDir, "Path to directory where pki files are stored.")
	fs.StringVarP(&o.namespace, "namespace", "n", o.namespace, "Name of Kubernetes namespace used to run guard server.")
	fs.StringVar(&o.addr, "addr", o.addr, "Address (host:port) of guard server.")
	fs.BoolVar(&o.runOnMaster, "run-on-master", o.runOnMaster, "If true, runs Guard server on master instances")
	fs.StringVar(&o.privateRegistry, "private-registry", o.privateRegistry, "Private Docker registry")
	fs.StringVar(&o.imagePullSecret, "image-pull-secret", o.imagePullSecret, "Name of image pull secret")
	o.Token.AddFlags(fs)
	o.Google.AddFlags(fs)
	o.Azure.AddFlags(fs)
	o.LDAP.AddFlags(fs)
}
