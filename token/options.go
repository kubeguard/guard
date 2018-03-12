package token

import (
	"github.com/spf13/pflag"
)

type Options struct {
	AuthFile string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.AuthFile, "token-auth-file", "", "To enable static token authentication")
}

func (o Options) ToArgs() []string {
	var args []string

	if o.AuthFile != "" {
		args = append(args, "--token-auth-file=/etc/guard/auth/token.csv")
	}

	return args
}

func (o *Options) Validate() []error {
	return nil
}
