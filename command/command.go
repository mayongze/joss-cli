package command

import (
	"fmt"
	"github.com/mayongze/joss-cli/pkg/joss"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

const (
	cliName        = "joss-cli"
	cliDescription = "A simple command line client for jdcloud oss."
)

var (
	RootCmd = &cobra.Command{
		Use:        cliName,
		Short:      cliDescription,
		Version:    Version,
		SuggestFor: []string{"joss-cli"},
	}

	DebugFlag   bool
	AccessKey   string
	SecretKey   string
	Region      string
	OssType     string
	JossType    = joss.OssTypeJDCloud
	Endpoint    string
	Internal    bool
	VersionFlag bool

	CommandTimeoutFlag string
	CommandTimeout     time.Duration
	Force              bool

	Recursive bool
)

func init() {
	// configure add
	cobra.EnablePrefixMatching = true
	cobra.OnInitialize(initConfig)
	RootCmd.SetVersionTemplate(version())

	RootCmd.PersistentFlags().StringVar(&AccessKey, "accessKey", "", "oss accessKey env: JOSS_ACCESSKEY")
	RootCmd.PersistentFlags().StringVar(&SecretKey, "secretKey", "", "oss secretKey env: JOSS_SECRETKEY")
	RootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "debug mode")
	RootCmd.PersistentFlags().StringVar(&Endpoint, "endpoint", "", "oss endpoint")
	RootCmd.PersistentFlags().StringVarP(&OssType, "oss-type", "t", "jd", "enable client-side debug logging")
	RootCmd.PersistentFlags().StringVar(&Region, "region", "cn-north-1", "region")
	RootCmd.PersistentFlags().BoolVar(&Internal, "internal", false, "is internal")
	RootCmd.PersistentFlags().StringVar(&CommandTimeoutFlag, "command-timeout", "5s", "timeout for short running command")
}

func initConfig() {
	// 设置各种默认值, 环境变量覆盖
	if DebugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	var err error
	CommandTimeout, err = time.ParseDuration(CommandTimeoutFlag)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "command-timeout flag error, uses the default value 5s")
		CommandTimeout, _ = time.ParseDuration("5s")
	}

	logrus.SetOutput(os.Stdout)
	// 环境变量读
	viper.SetEnvPrefix("JOSS")
	viper.RegisterAlias(joss.ACCESSKEY[0], "accessKey")
	viper.RegisterAlias(joss.SECRETKEY[0], "secretKey")
	_ = viper.BindEnv("ACCESSKEY")
	_ = viper.BindEnv("SECRETKEY")

	if AccessKey != "" && SecretKey != "" {
		viper.Set("accessKey", AccessKey)
		viper.Set("secretKey", SecretKey)
	} else {
		AccessKey = viper.GetString("accessKey")
		SecretKey = viper.GetString("secretKey")
	}

	joss.SetAkSk(AccessKey, SecretKey)
	if Endpoint != "" {
		joss.SetEndpoint(Endpoint)
	} else {
		switch OssType {
		case "jd":
			Endpoint = GetJDEndpoint(Internal)
			JossType = joss.OssTypeJDCloud
		}
	}
	// 目前不采用配置文件
	joss.SetEndpoint(Endpoint)
}

func GetJDEndpoint(internal bool) string {
	endpoint := ""
	switch Region {
	case "bj", "cn-north-1":
		Region = "cn-north-1"
	case "sh", "cn-east-2":
		Region = "cn-north-2"
	case "gz", "cn-south-1":
		Region = "cn-south-1"
	case "sq", "cn-east-1":
		Region = "cn-east-1"
	default:
		logrus.Panic("invalid region")
	}
	if !internal {
		endpoint = fmt.Sprintf("s3.%s.jdcloud-oss.com", Region)
	} else {
		endpoint = fmt.Sprintf("s3-internal.%s.jdcloud-oss.com", Region)
	}
	return endpoint
}
