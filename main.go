package main

import (
	"os"
	"fmt"

	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/guard/cmds"
)

func main() {
	asdfasfsadfasfd
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmds.NewRootCmd(Version).Execute(); err != nil {
		os.Exit(1)
	}
	fmt.Println("test")
	os.Exit(0)
}
