package cmds

import (
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/ntp"
	"github.com/appscode/guard/server"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	o := server.NewRecommendedOptions()
	srv := server.Server{
		RecommendedOptions: o,
	}
	maxClodkSkew := 5 * time.Second
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := ntp.CheckSkew(maxClodkSkew); err != nil {
				log.Fatal(err)
			}
			if !srv.RecommendedOptions.SecureServing.UseTLS() {
				log.Fatalln("Guard server must use SSL.")
			}
			srv.ListenAndServe()
		},
	}
	srv.AddFlags(cmd.Flags())
	cmd.Flags().DurationVar(&maxClodkSkew, "max-clock-skeew", maxClodkSkew, "Max acceptable clock skew for server clock")
	return cmd
}
