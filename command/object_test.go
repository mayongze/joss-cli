package command

import (
	"bytes"
	"testing"
)

func TestNewBucketObjectListCommand(t *testing.T) {
	output := new(bytes.Buffer)
	RootCmd.SetArgs([]string{"ls", "oss://joss-test/", "-l", "--max-keys=1"})
	RootCmd.SetOut(output)
	RootCmd.AddCommand(NewObjectListCommand())
	if err := RootCmd.Execute(); err != nil {
		t.Error("Unexpected error:", err)
	}
	t.Log(output.String())
}

func TestNewObjectPutCommand(t *testing.T) {
	output := new(bytes.Buffer)
	RootCmd.SetArgs([]string{"put", "../case", "oss://jcloud-opmid/josstest/aaaa", "-rf"})
	RootCmd.SetOut(output)
	RootCmd.AddCommand(NewObjectPutCommand())
	if err := RootCmd.Execute(); err != nil {
		t.Error("Unexpected error:", err)
	}
	t.Log(output.String())
}

func TestNewSignUrlCommand(t *testing.T) {
	t.Log(executeCmd("signurl", "oss://joss-tttt/backup/tttt.zip", "1h"))
}
