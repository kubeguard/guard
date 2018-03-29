package github

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Options struct {
	BaseUrl   string
	UploadUrl string
}

func NewOptions() Options {
	return Options{}
}

func (o *Options) Configure() error {
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.BaseUrl, "github.base-url", o.BaseUrl, "Base url for enterprise, keep empty to use default github base url")
	fs.StringVar(&o.UploadUrl, "github.upload-url", o.UploadUrl, "Upload url for enterprise, keep empty to use default github upload url")
}

func (o *Options) Validate() []error {
	var errs []error

	if o.BaseUrl != "" && o.UploadUrl == "" {
		errs = append(errs, errors.New("base url is provided, but upload url is empty"))
	}
	if o.BaseUrl == "" && o.UploadUrl != "" {
		errs = append(errs, errors.New("upload url is provided, but base url is empty"))
	}

	return errs
}

func (o Options) Apply(d *v1beta1.Deployment) (extraObjs []runtime.Object, err error) {
	args := d.Spec.Template.Spec.Containers[0].Args
	if o.BaseUrl != "" {
		args = append(args, fmt.Sprintf("--github.base-url=%s", o.BaseUrl))
	}
	if o.UploadUrl != "" {
		args = append(args, fmt.Sprintf("--github.upload-url=%s", o.UploadUrl))
	}

	d.Spec.Template.Spec.Containers[0].Args = args
	return extraObjs, nil
}
