package command

import (
	"bytes"
	"testing"
)

func TestNewBucketObjectListCommand(t *testing.T) {
	output := new(bytes.Buffer)

	testCmd.SetArgs([]string{"ls", "oss://joss-test", "-l"})
	testCmd.SetOutput(output)
	testCmd.AddCommand(NewBucketObjectListCommand())

	if err := testCmd.Execute(); err != nil {
		t.Error("Unexpected error:", err)
	}
	t.Log(output.String())
}
