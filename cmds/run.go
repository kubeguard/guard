package cmds

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/ntp"
	"github.com/appscode/guard/lib"
	"github.com/spf13/cobra"
)

const (
	servingPort = 8443
)

func NewCmdRun() *cobra.Command {
	srv := lib.Server{
		Address: fmt.Sprintf(":%d", servingPort),
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
			if !srv.UseTLS() {
				log.Fatalln("Guard server must use SSL.")
			}
			srv.ListenAndServe()
		},
	}
	srv.AddFlags(cmd.Flags())
	cmd.Flags().DurationVar(&maxClodkSkew, "max-clock-skeew", maxClodkSkew, "Max acceptable clock skew for server clock")
	return cmd
}
