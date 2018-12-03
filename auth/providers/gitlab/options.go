package gitlab

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/api/apps/v1beta1"
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

func (o Options) Apply(d *v1beta1.Deployment) (extraObjs []runtime.Object, err error) {
	args := d.Spec.Template.Spec.Containers[0].Args
	if o.BaseUrl != "" {
		args = append(args, fmt.Sprintf("--gitlab.base-url=%s", o.BaseUrl))
	}
	args = append(args, fmt.Sprintf("--gitlab.use-group-id=%t", o.UseGroupID))

	d.Spec.Template.Spec.Containers[0].Args = args

	return extraObjs, nil
}
