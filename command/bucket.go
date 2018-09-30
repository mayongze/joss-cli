package command

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

func NewBucketCommand() *cobra.Command {
	bc := &cobra.Command{
		Use:   "bucket <subcommand>",
		Short: "Bucket related commands",
	}
	bc.AddCommand(NewBucketAddCommand())
	bc.AddCommand(NewBucketRemoveCommand())
	bc.AddCommand(NewBucketListCommand())
	return bc
}

func NewBucketAddCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "add <BucketName>",
		Short: "Adds a bucket into the oss",

		Run: bucketAddCommandFunc,
	}
	return c
}

func NewBucketRemoveCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "remove <BucketName>",
		Short: "Removes a bucket from the oss",

		Run: bucketRemoveCommandFunc,
	}
	return c
}

func NewBucketListCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "ls",
		Aliases: []string{"list"},
		Short: "获取bucket列表.",

		Run: bucketListCommandFunc,
	}
	c.Flags().BoolP("long","l",false,"显示详细信息.")
	return c
}


func bucketAddCommandFunc(cmd *cobra.Command, args []string) {

}

func bucketRemoveCommandFunc(cmd *cobra.Command, args []string) {

}

func bucketListCommandFunc(cmd *cobra.Command, args []string) {
	s3client,_ := NewS3Client(cmd)
	resp, err := s3client.ListBuckets(&s3.ListBucketsInput{})

	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR] oss list bucket failed. %v", err)
	}

	//输出格式
	longMode,_ := cmd.Flags().GetBool("long")
	if longMode {
		fmt.Fprintf(os.Stdout,"total %n\n",len(resp.Buckets))
		for _, bucket := range resp.Buckets {
			fmt.Fprintf(os.Stdout,"%s\t%s\n",bucket.CreationDate.Format("2006-01-02 15:04"), bucket.String())
		}
	}else{
		for _, bucket := range resp.Buckets {
			fmt.Fprintf(os.Stdout,"%s\t",*bucket.Name)
		}
		fmt.Fprint(os.Stdout,"\n")
	}
}

//获取bucket 内的文件
func NewBucketObjectListCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "ls oss://BUCKET[/PREFIX]",
		Aliases: []string{"list"},
		Short: "获取oss里的列表信息,如需获取对象信息需要使用oss://指定bucket",

		Run:ossListCommandFunc,
	}
	c.Flags().BoolP("long","l",false,"显示详细信息.")
	c.Flags().Int64("max-keys",0,"keys的最大输出数目.")
	return c
}

//获取对象存储列表
func ossListCommandFunc(cmd *cobra.Command, args []string){
	//无附加参数获取bucket 列表
	if len(args) == 0{
		bucketListCommandFunc(cmd,args)
		return
	}
	//解析参数  oss://bucket/aae/f
	param := args[0]
	bucket := ""
	prefix := ""
	if strings.HasPrefix(param,"oss://") {
		idx := strings.LastIndex(param[6:],"/")
		if idx <= 0{
			bucket = param[6:]
		}else{
			bucket = param[6:idx]
			prefix = param[6+idx:]
		}
	}else{
		ExitWithError(ExitError,fmt.Errorf("bucket格式错误. Format: ls oss://bucket[/prefix]"))
	}

	if bucket == ""{
		ExitWithError(ExitError,fmt.Errorf("bucket不能为空. Format: ls oss://bucket[/prefix]"))
	}

	s3cli,_ := NewS3Client(cmd)
	maxKeys,_ := cmd.Flags().GetInt64("max-keys")

	objects,err := getListObject(s3cli,bucket,prefix,maxKeys)
	if err != nil{
		ExitWithError(ExitError,err)
	}

	longMode,_ := cmd.Flags().GetBool("long")
	if longMode == true {
		 fmt.Fprintf(os.Stdout,"total %d\n",len(objects))
		 for _,obj :=  range objects{
			//|s|%s %3d %s|s|%8s|%s %-7s %s %-3s
		 	fmt.Fprintf(os.Stdout,"%s\t%9d\t%6s\t%s\n",
		 		*obj.Owner.DisplayName,
				*obj.Size,
		 		obj.LastModified.Format("2006-01-02 15:04"),
		 		*obj.Key)
		 }

	}else{
		for _,obj :=  range objects{
			fmt.Fprintf(os.Stdout,"%s\t",*obj.Key)
		}
		fmt.Fprintf(os.Stdout,"\n")
	}
}

//获取对象存储列表
func getListObject(s3cli *s3.S3,bucket,prefix string,maxKeys int64) ( objects []*s3.Object,err error){
	//超时时间30秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	var pMaxKeys *int64 = nil
	if maxKeys > 0{
		pMaxKeys = aws.Int64(maxKeys)
	}

	out, err := s3cli.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
		MaxKeys: pMaxKeys,
	})

	if err != nil{
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				err = fmt.Errorf("bucket不存在.")
			case request.CanceledErrorCode:
				err = fmt.Errorf("获取对象列表超时,%s",aerr.Error())
			default:
				err = fmt.Errorf(aerr.Error())
			}
		}
		return
	}
	objects = out.Contents
	return objects,nil
}
