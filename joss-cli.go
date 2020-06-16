package main

import (
	"fmt"
	"github.com/mayongze/joss-cli/command"
	"os"
)

func main() {
	//命令执行
	if err := command.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
