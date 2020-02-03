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
package firebase

import (
	"io/ioutil"

	"github.com/appscode/go/types"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Options struct {
	ServiceAccountJsonFile string
}

func NewOptions() Options {
	return Options{}
}

func (o *Options) Configure() error {
	if o.ServiceAccountJsonFile != "" {
		_, err := ioutil.ReadFile(o.ServiceAccountJsonFile)
		if err != nil {
			return errors.Wrapf(err, "failed to load service account json file %s", o.ServiceAccountJsonFile)
		}
	}

	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ServiceAccountJsonFile, "firebase.sa-json-file", o.ServiceAccountJsonFile, "Path to Google service account json file")
}

func (o *Options) Validate() []error {
	var errs []error
	if o.ServiceAccountJsonFile == "" {
		errs = append(errs, errors.New("firebase.sa-json-file must be non-empty"))
	}
	return errs
}

func (o Options) Apply(d *apps.Deployment) (extraObjs []runtime.Object, err error) {
	container := d.Spec.Template.Spec.Containers[0]

	// create auth secret
	sa, err := ioutil.ReadFile(o.ServiceAccountJsonFile)
	if err != nil {
		return nil, err
	}
	authSecret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-firebase-auth",
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
		MountPath: "/etc/guard/auth/firebase",
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

	// export GOOGLE_APPLICATION_CREDENTIALS
	// https://cloud.google.com/docs/authentication/getting-started
	envs := container.Env
	if o.ServiceAccountJsonFile != "" {
		e := core.EnvVar{
			Name:  "GOOGLE_APPLICATION_CREDENTIALS",
			Value: "/etc/guard/auth/firebase/sa.json",
		}
		envs = append(envs, e)
	}
	container.Env = envs

	d.Spec.Template.Spec.Containers[0] = container

	return extraObjs, nil
}
