package main

import (
	"os"

	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/guard/commands"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := commands.NewRootCmd(Version).Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
