package google

import (
	"fmt"
	"io/ioutil"

	"github.com/appscode/go/types"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	gdir "google.golang.org/api/admin/directory/v1"
	"k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Options struct {
	ServiceAccountJsonFile string
	AdminEmail             string
	jwtConfig              *jwt.Config
}

func NewOptions() Options {
	return Options{}
}

func (o *Options) Bootstrap() error {
	if o.ServiceAccountJsonFile != "" {
		sa, err := ioutil.ReadFile(o.ServiceAccountJsonFile)
		if err != nil {
			return errors.Wrapf(err, "failed to load service account json file %s", o.ServiceAccountJsonFile)
		}

		o.jwtConfig, err = google.JWTConfigFromJSON(sa, gdir.AdminDirectoryGroupReadonlyScope)
		if err != nil {
			return errors.Wrapf(err, "failed to create JWT config from service account json file %s", o.ServiceAccountJsonFile)
		}

		// https://admin.google.com/ManageOauthClients
		// ref: https://developers.google.com/admin-sdk/directory/v1/guides/delegation
		// Note: Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore your service account needs to impersonate one of those users to access the Admin SDK Directory API.
		o.jwtConfig.Subject = o.AdminEmail
	}

	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ServiceAccountJsonFile, "google.sa-json-file", o.ServiceAccountJsonFile, "Path to Google service account json file")
	fs.StringVar(&o.AdminEmail, "google.admin-email", o.AdminEmail, "Email of G Suite administrator")
}

func (o *Options) Validate() []error {
	return nil
}

func (o Options) IsSet() bool {
	return o.ServiceAccountJsonFile != ""
}

func (o Options) Apply(d *v1beta1.Deployment) (extraObjs []runtime.Object, err error) {
	if !o.IsSet() {
		return nil, nil // nothing to apply
	}

	container := d.Spec.Template.Spec.Containers[0]

	// create auth secret
	sa, err := ioutil.ReadFile(o.ServiceAccountJsonFile)
	if err != nil {
		return nil, err
	}
	authSecret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-google-auth",
			Namespace: d.Namespace,
			Labels:    d.Labels,
		},
		Data: map[string][]byte{
			"sa.json": sa,
		},
	}
	extraObjs = append(extraObjs, authSecret)

	// mount auth secret into deployment
	volMount := core.VolumeMount{
		Name:      authSecret.Name,
		MountPath: "/etc/guard/auth/google",
	}
	container.VolumeMounts = append(container.VolumeMounts, volMount)

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
	args := container.Args
	if o.ServiceAccountJsonFile != "" {
		args = append(args, "--google.sa-json-file=/etc/guard/auth/google/sa.json")
	}
	if o.AdminEmail != "" {
		args = append(args, fmt.Sprintf("--google.admin-email=%s", o.AdminEmail))
	}

	return extraObjs, nil
}
