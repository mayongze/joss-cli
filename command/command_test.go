package command

import (
	"testing"
)

func Test_Init(t *testing.T) {
	RootCmd.SetArgs([]string{"bucket", "ls", "-l"})
	_ = RootCmd.Execute()
}
