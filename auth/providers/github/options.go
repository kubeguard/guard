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
package github

import (
	"fmt"

	"github.com/spf13/pflag"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Options struct {
	BaseUrl string
}

func NewOptions() Options {
	return Options{}
}

func (o *Options) Configure() error {
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.BaseUrl, "github.base-url", o.BaseUrl, "Base url for enterprise, keep empty to use default github base url")
}

func (o *Options) Validate() []error {
	return nil
}

func (o Options) Apply(d *apps.Deployment) (extraObjs []runtime.Object, err error) {
	args := d.Spec.Template.Spec.Containers[0].Args
	if o.BaseUrl != "" {
		args = append(args, fmt.Sprintf("--github.base-url=%s", o.BaseUrl))
	}

	d.Spec.Template.Spec.Containers[0].Args = args
	return extraObjs, nil
}
