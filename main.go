package main

import (
	"os"

	"github.com/appscode/kad/cmds"
	logs "github.com/appscode/log/golog"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmds.NewRootCmd(Version).Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
