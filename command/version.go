package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	Version = "1.0.0+git"
	// We want to replace this variable at build time with "-ldflags -X main.GitSHA=xxx", where const is not supported.
	GitSHA = ""
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the version of joss",
		Run:   versionCommandFunc,
	}
}

func versionCommandFunc(cmd *cobra.Command, args []string) {
	fmt.Printf("joss-cli version:%s sha:%s\n", Version, GitSHA)
}
