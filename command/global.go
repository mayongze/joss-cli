package command

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	"strings"
	_ "time"
)

var (
	ShellRootPath string
	AccountFile   string
)

type GlobalFlags struct {
	Endpoint  string
	AccessKey string
	SecretKey string

	RegionId string

	//Timeout time.Duration

	Debug bool
}

func NewS3Client(cmd *cobra.Command) (*s3.S3, *session.Session) {
	accessKey, _ := cmd.Flags().GetString("accessKey")
	secretKey, _ := cmd.Flags().GetString("secretKey")
	endpoint, _ := cmd.Flags().GetString("endpoint")
	if accessKey == "" || secretKey == "" {
		account, err := GetAcount()
		if err != nil {
			ExitWithError(ExitError, fmt.Errorf("Access key error. %v", err))
		}
		accessKey = account.AccessKey
		secretKey = account.SecretKey
	}
	return initS3Client(accessKey, secretKey, endpoint)
}

func initS3Client(accessKey, secretKey, endpoint string) (*s3.S3, *session.Session) {
	if !strings.HasPrefix(endpoint, "http://") {
		endpoint = "http://" + endpoint
	}
	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint:    aws.String(endpoint),
			Region:      aws.String("oss"),
			Credentials: creds,
			DisableSSL:  aws.Bool(true),
		},
		Profile: "jdcloud",
	}))
	svc := s3.New(sess)
	return svc, sess
}
