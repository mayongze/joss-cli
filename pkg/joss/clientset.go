package joss

import (
	"context"
	"errors"
	"github.com/mayongze/joss-cli/pkg/joss/awss3"
	"github.com/mayongze/joss-cli/pkg/joss/types"
	"io/ioutil"
	"os"
)

type OssType string

const (
	OssTypeJDCloud OssType = "JDCloud"
	OssTypeS3              = "s3"
	OssTypeAzure           = "Azure"
	OssTypeAliYun          = "Aliyun"
)

type Clientset struct {
	endpoint  string
	accessKey string
	secretKey string

	t OssType
	types.ClientInterface
}

func New(endpoint, ak, sk string, t OssType) *Clientset {
	// 返回工厂模式
	var c types.ClientInterface
	switch t {
	case OssTypeS3, OssTypeJDCloud:
		s3c, _ := initS3Client(ak, sk, endpoint, false)
		c = awss3.New(s3c)
	default:
		panic("暂不支持")
	}
	return &Clientset{
		endpoint:        endpoint,
		accessKey:       ak,
		secretKey:       sk,
		t:               t,
		ClientInterface: c,
	}
}

func (cs *Clientset) GetBytesObject(ctx context.Context, path string) (b []byte, err error) {
	// 获取path
	bucket, objectKey, err := ParseBucketAndKey(path)
	if err != nil {
		return nil, err
	}
	reader, size, err := cs.Bucket(bucket).GetObject(objectKey, types.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	b, err = ioutil.ReadAll(reader)
	if len(b) != int(size) {
		err = errors.New("size don't match")
	}
	return
}

func (cs *Clientset) GetStringObject(ctx context.Context, path string) (string, error) {
	b, err := cs.GetBytesObject(ctx, path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (cs *Clientset) GetObjectToFile(objectPath string, filePath string) error {
	// 文件夹递归? -r

	return nil
}

func (cs *Clientset) PutObjectFromFile(objectPath string, filePath string, options ...types.OpOption) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()
	bucket, key, err := ParseBucketAndKey(objectPath)
	if err != nil {
		return err
	}

	if err = cs.Bucket(bucket).PutObject(key, fd, options...); err != nil {
		return err
	}
	return nil
}
