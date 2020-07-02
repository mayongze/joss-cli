package joss

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"strings"
)

var (
	ENDPOINT = []string{"endpoint"}

	ACCESSKEY = []string{"access_key"}
	SECRETKEY = []string{"secret_key"}
)

func Endpoint() string {
	endpoint := viper.GetString(ENDPOINT[0])
	return endpoint
}

func AccessKey() string {
	accesskey := viper.GetString(ACCESSKEY[0])
	return accesskey
}

func SecretKey() string {
	return viper.GetString(SECRETKEY[0])
}

func SetAkSk(ak, sk string) {
	viper.Set(ACCESSKEY[0], ak)
	viper.Set(SECRETKEY[0], sk)
}

func SetEndpoint(endpoint string) {
	viper.Set(ENDPOINT[0], endpoint)
}

func NewS3Client() (*s3.S3, *session.Session) {
	account, err := GetAccount()
	if err != nil {
		panic(err)
	}
	endpoint := Endpoint()
	return initS3Client(account.AccessKey, account.SecretKey, endpoint, true)
}

func initS3Client(accessKey, secretKey, endpoint string, forcePath bool) (*s3.S3, *session.Session) {
	if !strings.HasPrefix(endpoint, "http://") {
		endpoint = "http://" + endpoint
	}

	var creds *credentials.Credentials
	if accessKey == "" {
		creds = credentials.AnonymousCredentials
	} else {
		creds = credentials.NewStaticCredentials(accessKey, secretKey, "")
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint:         aws.String(endpoint),
			Region:           aws.String("oss"),
			Credentials:      creds,
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(forcePath),
		},
	}))
	svc := s3.New(sess)
	return svc, sess
}
