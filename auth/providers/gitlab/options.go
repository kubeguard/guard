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

package gitlab

import (
	"fmt"

	"github.com/spf13/pflag"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Options struct {
	BaseUrl    string
	UseGroupID bool
}

func NewOptions() Options {
	return Options{
		UseGroupID: false,
	}
}

func (o *Options) Configure() error {
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.BaseUrl, "gitlab.base-url", o.BaseUrl, "Base url for GitLab, including the API path, keep empty to use default gitlab base url.")
	fs.BoolVar(&o.UseGroupID, "gitlab.use-group-id", o.UseGroupID, "Use group ID for authentication instead of group full path")
}

func (o *Options) Validate() []error {
	return nil
}

func (o Options) Apply(d *apps.Deployment) (extraObjs []runtime.Object, err error) {
	args := d.Spec.Template.Spec.Containers[0].Args
	if o.BaseUrl != "" {
		args = append(args, fmt.Sprintf("--gitlab.base-url=%s", o.BaseUrl))
	}
	args = append(args, fmt.Sprintf("--gitlab.use-group-id=%t", o.UseGroupID))

	d.Spec.Template.Spec.Containers[0].Args = args

	return extraObjs, nil
}
