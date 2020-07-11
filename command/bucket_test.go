package command

import (
	"bytes"
	"testing"
)

func executeCmd(args ...string) string {
	output := new(bytes.Buffer)
	RootCmd.SetArgs(args)
	RootCmd.SetOut(output)
	_ = RootCmd.Execute()
	return output.String()
}

func TestNewBucketListCommand(t *testing.T) {
	t.Log(executeCmd("bucket", "ll"))
}

func TestNewBucketAddCommand(t *testing.T) {
	t.Log(executeCmd("bucket", "add", "josstest"))
	t.Log(executeCmd("bucket", "remove", "josstest"))
}
