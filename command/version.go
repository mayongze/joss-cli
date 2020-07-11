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

func init() {
	RootCmd.AddCommand(NewVersionCommand())
}

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the version of joss",
		Run:   versionCommandFunc,
	}
}

func versionCommandFunc(cmd *cobra.Command, args []string) {
	fmt.Println(version())
}

func version() string {
	return fmt.Sprintf("version: %s \nsha: %s\n", Version, GitSHA)
}
