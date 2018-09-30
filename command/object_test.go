package command

import (
	"bytes"
	"github.com/spf13/cobra"
	"testing"
)


func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOutput(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func TestNewObjectPutCommand(t *testing.T) {
	output := new(bytes.Buffer)
	path := "D:\\xmind-8-update8-windows.zip"

	testCmd.SetArgs([]string{"put",path,"oss://joss-test"})
	testCmd.SetOutput(output)
	testCmd.AddCommand(NewObjectPutCommand())

	if err := testCmd.Execute(); err != nil {
		t.Error("Unexpected error:", err)
	}
	t.Log(output.String())
}

func TestGetFileObject(t *testing.T) {
	path := []string{"../case/../command","../case/../../../../"}
	//../command ../../../../
	fileObj,err := getFileObject(path)
	if err != nil {
		t.Error("getFileObject error:%v",err)

	}
	t.Log(fileObj)
}