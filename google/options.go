package google

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Options struct {
	ServiceAccountJsonFile string
	AdminEmail             string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ServiceAccountJsonFile, "google.sa-json-file", o.ServiceAccountJsonFile, "Path to Google service account json file")
	fs.StringVar(&o.AdminEmail, "google.admin-email", o.AdminEmail, "Email of G Suite administrator")
}

func (o Options) ToArgs() []string {
	var args []string

	if o.ServiceAccountJsonFile != "" {
		args = append(args, "--google.sa-json-file=/etc/guard/auth/sa.json")
	}
	if o.AdminEmail != "" {
		args = append(args, fmt.Sprintf("--google.admin-email=%s", o.AdminEmail))
	}

	return args
}

func (o *Options) Validate() []error {
	return nil
}
