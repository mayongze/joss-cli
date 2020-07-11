package types

import (
	"io"
	"time"
)

type ClientInterface interface {
	Bucket(string) BucketMgrInterface
	CreateBucket(string) (BucketMgrInterface, error)
	DeleteBucket(string) error
	ListBucket() (ListBucketResp, error)
}

type BucketMgrInterface interface {
	PutObject(objectKey string, reader io.Reader, options ...OpOption) error
	GetObject(objectKey string, option ...OpOption) (io.ReadCloser, int64, error)
	DeleteObjects(objectKeys []string, options ...OpOption) (DeleteObjectResp, error)
	// 数量
	ListObject(keyPrefix string, options ...OpOption) (ListObjectResp, error)
	// get sign url
	GetObjectSignUrl(key string, expire time.Duration, options ...OpOption) (string, error)
}

type (
	ListBucketResp struct {
		Bucket []Bucket
	}
	Bucket struct {
		CreationDate time.Time
		Name         string
	}

	ListObjectResp struct {
		Prefix  []string
		Objects []Object
	}

	Object struct {
		ETag         string
		Key          string
		Size         int64
		LastModified time.Time
	}

	DeleteObjectResp struct {
		Keys []string
	}
)
