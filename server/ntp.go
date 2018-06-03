package server

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

type NTPOptions struct {
	MaxClodkSkew time.Duration
	Interval     time.Duration
}

func NewNTPOptions() NTPOptions {
	return NTPOptions{
		MaxClodkSkew: 2 * time.Minute,
		Interval:     10 * time.Minute,
	}
}

func (o *NTPOptions) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&o.MaxClodkSkew, "max-clock-skew", o.MaxClodkSkew, "Max acceptable clock skew for server clock")
	fs.DurationVar(&o.Interval, "clock-check-interval", o.Interval, "Interval between checking time against NTP servers")
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
