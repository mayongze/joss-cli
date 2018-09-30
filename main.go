package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"joss-cli/command"
	"os"
	_ "time"
)

const (
	cliName        = "joss-cli"
	cliDescription = "A simple command line client for jdcloud oss."
	//accessKey      = ""
	//secretKey      = ""
)

var (
	globalFlags = command.GlobalFlags{}
	rootCmd     = &cobra.Command{
		Use:        cliName,
		Short:      cliDescription,
		SuggestFor: []string{"joss-cli"},
	}
)

func init() {
	// configure add
	cobra.EnablePrefixMatching = true
	rootCmd.PersistentFlags().StringVar(&globalFlags.Endpoint, "endpoint", "s3.cn-north-1.jcloudcs.com", "oss endpoint")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "enable client-side debug logging")
	rootCmd.PersistentFlags().StringVar(&globalFlags.AccessKey, "accessKey", "", "jcloud accessKey")
	rootCmd.PersistentFlags().StringVar(&globalFlags.SecretKey, "secretKey", "", "jcloud accessKey")
	rootCmd.PersistentFlags().StringVar(&globalFlags.RegionId, "regionId", "cn-north-1", "regionId")

	rootCmd.AddCommand(command.NewBucketCommand())
	rootCmd.AddCommand(command.NewObjectPutCommand())
	rootCmd.AddCommand(command.NewBucketObjectListCommand())
	rootCmd.AddCommand(command.NewAccountCommand())
	rootCmd.AddCommand(command.NewVersionCommand())
}

func main() {
	//命令执行
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
