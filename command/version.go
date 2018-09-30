package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (

	Version           = "1.0.0+git"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the version of joss",
		Run:   versionCommandFunc,
	}
}

func versionCommandFunc(cmd *cobra.Command, args []string) {
	fmt.Println("joss-cli version:", Version)
}

