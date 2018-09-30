package command

import "github.com/spf13/cobra"

var (

	testCmd = &cobra.Command{}

	endpoint  = "s3.cn-north-1.jcloudcs.com"
	accessKey = ""
	secretKey = ""
)

func init(){
	testCmd.PersistentFlags().String("endpoint", endpoint, "oss endpoint")
	testCmd.PersistentFlags().Bool("debug", false, "enable client-side debug logging")
	testCmd.PersistentFlags().String("accessKey", accessKey, "jcloud accessKey")
	testCmd.PersistentFlags().String("secretKey", secretKey, "jcloud accessKey")
	testCmd.PersistentFlags().String("regionId", "cn-north-222", "regionId")
}