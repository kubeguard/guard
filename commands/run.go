package commands

import (
	"github.com/appscode/go/log"
	"github.com/appscode/guard/server"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	o := server.NewRecommendedOptions()
	srv := server.Server{
		RecommendedOptions: o,
	}
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if !srv.RecommendedOptions.SecureServing.UseTLS() {
				log.Fatalln("Guard server must use SSL.")
			}
			srv.ListenAndServe()
		},
	}
	srv.AddFlags(cmd.Flags())
	return cmd
}
