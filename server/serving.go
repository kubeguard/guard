package server

import (
	"fmt"
	"path/filepath"

	"github.com/appscode/go/types"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"gomodules.xyz/cert/certstore"
	"k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ServingPort = 8443
)

type SecureServingOptions struct {
	SecureAddr string
	CACertFile string
	CertFile   string
	KeyFile    string

	pkiDir string
}

func NewSecureServingOptions() SecureServingOptions {
	return SecureServingOptions{
		SecureAddr: fmt.Sprintf(":%d", ServingPort),
	}
}

func NewSecureServingOptionsFromDir(pkiDir string) SecureServingOptions {
	return SecureServingOptions{
		pkiDir: pkiDir,
	}
}

func (o *SecureServingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.SecureAddr, "secure-addr", o.SecureAddr, "host:port used to serve secure apis")
	fs.StringVar(&o.CACertFile, "tls-ca-file", o.CACertFile, "File containing CA certificate")
	fs.StringVar(&o.CertFile, "tls-cert-file", o.CertFile, "File container server TLS certificate")
	fs.StringVar(&o.KeyFile, "tls-private-key-file", o.KeyFile, "File containing server TLS private key")
}

func (o SecureServingOptions) UseTLS() bool {
	return o.CACertFile != "" && o.CertFile != "" && o.KeyFile != ""
}

func (o *SecureServingOptions) Validate() []error {
	var errs []error
	if o.SecureAddr == "" {
		errs = append(errs, errors.New("server address is empty"))
	}
	if o.CACertFile == "" {
		errs = append(errs, errors.New("CA cert is empty"))
	}
	if o.CACertFile == "" {
		errs = append(errs, errors.New("CA cert is empty"))
	}
	if o.CertFile == "" {
		errs = append(errs, errors.New("server certificate is empty"))
	}
	if o.KeyFile == "" {
		errs = append(errs, errors.New("server key is empty"))
	}
	return errs
}

func (o SecureServingOptions) Apply(d *v1beta1.Deployment) (extraObjs []runtime.Object, err error) {
	// create auth secret
	store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(o.pkiDir, "pki"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate store.")
	}
	if !store.PairExists("ca") {
		return nil, errors.Errorf("CA certificates not found in %s. Run `guard init ca`", store.Location())
	}
	if !store.PairExists("server") {
		return nil, errors.Errorf("Server certificate not found in %s. Run `guard init server`", store.Location())
	}

	caCert, _, err := store.ReadBytes("ca")
	if err != nil {
		return nil, errors.Wrap(err, "failed to load ca certificate.")
	}
	serverCert, serverKey, err := store.ReadBytes("server")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load ca certificate.")
	}

	authSecret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-pki",
			Namespace: d.Namespace,
			Labels:    d.Labels,
		},
		Data: map[string][]byte{
			"ca.crt":  caCert,
			"tls.crt": serverCert,
			"tls.key": serverKey,
		},
	}
	extraObjs = append(extraObjs, authSecret)

	// mount auth secret into deployment
	volMount := core.VolumeMount{
		Name:      authSecret.Name,
		MountPath: "/etc/guard/pki",
	}
	d.Spec.Template.Spec.Containers[0].VolumeMounts = append(d.Spec.Template.Spec.Containers[0].VolumeMounts, volMount)

	vol := core.Volume{
		Name: authSecret.Name,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName:  authSecret.Name,
				DefaultMode: types.Int32P(0555),
			},
		},
	}
	d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, vol)

	// use auth secret in container[0] args
	args := d.Spec.Template.Spec.Containers[0].Args
	args = append(args, "--tls-ca-file=/etc/guard/pki/ca.crt")
	args = append(args, "--tls-cert-file=/etc/guard/pki/tls.crt")
	args = append(args, "--tls-private-key-file=/etc/guard/pki/tls.key")
	d.Spec.Template.Spec.Containers[0].Args = args

	return extraObjs, nil
}
