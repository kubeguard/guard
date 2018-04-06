package commands

import (
	"fmt"

	"github.com/appscode/guard/installer"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func NewCmdInstaller() *cobra.Command {
	opts := installer.New()
	cmd := &cobra.Command{
		Use:               "installer",
		Short:             "Prints Kubernetes objects for deploying guard server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			errs := opts.Validate()
			if errs != nil {
				glog.Fatal(errs)
			}

			data, err := installer.Generate(opts)
			if err != nil {
				glog.Fatal(err)
			}
			fmt.Println(string(data))
		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}
