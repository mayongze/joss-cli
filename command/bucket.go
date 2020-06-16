package command

import (
	"fmt"
	"github.com/mayongze/joss-cli/pkg/joss"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	bucketCmd := &cobra.Command{
		Use:   "bucket <SubCommand>",
		Short: "Bucket related commands",
	}
	bucketCmd.AddCommand(
		NewBucketAddCommand(),
		NewBucketRemoveCommand(),
		NewBucketListCommand())

	RootCmd.AddCommand(bucketCmd)
}

func NewBucketAddCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "add <BucketName>",
		Short: "添加一个Bucket存储桶",
		Run:   bucketAddCommandFunc,
		Args:  cobra.ExactArgs(1),
	}

	return c
}

func NewBucketRemoveCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "remove <BucketName>",
		Short: "删除一个Bucket存储桶",
		Args:  cobra.ExactArgs(1),
		Run:   bucketRemoveCommandFunc,
	}
	return c
}

func NewBucketListCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list", "ll"},
		Short:   "获取bucket列表",

		Run: bucketListCommandFunc,
	}
	c.Flags().BoolP("long", "l", false, "显示详细信息")
	return c
}

func bucketAddCommandFunc(cmd *cobra.Command, args []string) {
	bucketName := args[0]
	jcli := joss.New(Endpoint, AccessKey, SecretKey, JossType)
	resp, err := jcli.CreateBucket(bucketName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERR] oss add bucket failed.\n %v", err)
		return
	}
	_ = resp
	fmt.Println(bucketName)
}

func bucketRemoveCommandFunc(cmd *cobra.Command, args []string) {
	bucketName := args[0]
	jcli := joss.New(Endpoint, AccessKey, SecretKey, JossType)
	err := jcli.DeleteBucket(bucketName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERR] oss add bucket failed. %v", err)
		return
	}
	fmt.Printf("删除bucket:%s成功!", bucketName)
}

// 获取oss内所有的存储桶
func bucketListCommandFunc(cmd *cobra.Command, args []string) {
	jcli := joss.New(Endpoint, AccessKey, SecretKey, JossType)
	resp, err := jcli.ListBucket()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERR] oss list bucket failed. %v", err)
		return
	}
	//输出格式
	longMode, _ := cmd.Flags().GetBool("long")
	if longMode || cmd.CalledAs() == "ll" {
		_, _ = fmt.Fprintf(os.Stdout, "total %d\n", len(resp.Bucket))
		for _, bucket := range resp.Bucket {
			_, _ = fmt.Fprintf(os.Stdout, "%s\toss://%s\n", bucket.CreationDate.Format("2006-01-02 15:04"), bucket.Name)
		}
	} else {
		for _, bucket := range resp.Bucket {
			_, _ = fmt.Fprintf(os.Stdout, "%s\t", bucket.Name)
		}
		_, _ = fmt.Fprint(os.Stdout, "\n")
	}
}
