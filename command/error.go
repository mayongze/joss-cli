package command

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
)

const (
	// http://tldp.org/LDP/abs/html/exitcodes.html
	ExitSuccess = iota
	ExitError
	ExitBadConnection
	ExitInvalidInput // for txn, watch command
	ExitBadFeature   // provided a valid flag with an unsupported value
	ExitInterrupted
	ExitIO
	ExitBadArgs = 128
)

func ExitWithError(code int, err error) {
	BucketError(err)

	fmt.Fprintln(os.Stderr, "Error:", err)
	/*
		if cerr, ok := err.(*client.ClusterError); ok {
			fmt.Fprintln(os.Stderr, cerr.Detail())
		}*/
	os.Exit(code)
}

func ExitWithBucketError(err error) {
	BucketError(err)
	os.Exit(1)
}

func BucketError(err error) error {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case s3.ErrCodeBucketAlreadyExists:
			fmt.Fprintln(os.Stderr, s3.ErrCodeBucketAlreadyExists, aerr.Error())
		case s3.ErrCodeBucketAlreadyOwnedByYou:
			fmt.Fprintln(os.Stderr, s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
		case s3.ErrCodeNoSuchBucket:
			fmt.Fprintln(os.Stderr, s3.ErrCodeNoSuchBucket, aerr.Error())
		default:
			fmt.Fprintln(os.Stderr, aerr.Error())
		}
	}
	return err
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
