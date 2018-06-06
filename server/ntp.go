package server

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

type NTPOptions struct {
	NTPServer    string
	MaxClodkSkew time.Duration
	Interval     time.Duration
}

func NewNTPOptions() NTPOptions {
	return NTPOptions{
		NTPServer:    "0.pool.ntp.org",
		MaxClodkSkew: 2 * time.Minute,
		Interval:     10*time.Minute,
	}
}

func (o *NTPOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.NTPServer, "ntp-server", o.NTPServer, "Address of NTP serer used to check clock skew")
	fs.DurationVar(&o.MaxClodkSkew, "max-clock-skew", o.MaxClodkSkew, "Max acceptable clock skew for server clock")
	fs.DurationVar(&o.Interval, "clock-check-interval", o.Interval, "Interval between checking time against NTP servers, set to 0 to disable checks")
}

func (o NTPOptions) ToArgs() []string {
	var args []string

	if o.MaxClodkSkew > 0 {
		args = append(args, fmt.Sprintf("--max-clock-skew=%v", o.MaxClodkSkew))
	}
	if o.Interval > 0 {
		args = append(args, fmt.Sprintf("--clock-check-interval=%v", o.Interval))
	}

	return args
}

func (o *NTPOptions) Validate() []error {
	return nil
}

func (o *NTPOptions) Enabled() bool {
	return o.Interval > 0
}
