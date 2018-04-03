package commands

import (
	"fmt"
	"log"

	"github.com/appscode/guard/installer"
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
				log.Fatal(errs)
			}

			data, err := installer.Generate(opts)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(data))
		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}
