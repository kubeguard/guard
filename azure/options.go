package azure

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Options struct {
	ClientID     string
	ClientSecret string
	TenantID     string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ClientID, "azure.client-id", o.ClientID, "MS Graph application client ID to use")
	fs.StringVar(&o.ClientSecret, "azure.client-secret", o.ClientSecret, "MS Graph application client secret to use")
	fs.StringVar(&o.TenantID, "azure.tenant-id", o.TenantID, "MS Graph application tenant id to use")
}

func (o Options) ToArgs() []string {
	var args []string

	if o.ClientID != "" {
		args = append(args, fmt.Sprintf("--azure.client-id=%s", o.ClientID))
	}
	if o.ClientSecret != "" {
		args = append(args, fmt.Sprintf("--azure.client-secret=%s", o.ClientSecret))
	}
	if o.TenantID != "" {
		args = append(args, fmt.Sprintf("--azure.tenant-id=%s", o.TenantID))
	}

	return args
}

func (o *Options) Validate() []error {
	return nil
}
