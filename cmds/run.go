package cmds

import (
	"fmt"

	"github.com/appscode/go/log"
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
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if !srv.UseTLS() {
				log.Fatalln("Guard server must use SSL.")
			}
			srv.ListenAndServe()
		},
	}
	srv.AddFlags(cmd.Flags())
	return cmd
}
