package awss3

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mayongze/joss-cli/pkg/joss/types"
	"io"
	"net/url"
	"time"
)

type S3Mgr struct {
	*s3.S3
}

func New(s3cli *s3.S3) *S3Mgr {
	v, _ := s3cli.Config.Credentials.Get()
	if !v.HasKeys() {
		s3cli.Config.Credentials = credentials.AnonymousCredentials
	}
	return &S3Mgr{s3cli}
}

func (s *S3Mgr) DeleteBucket(bucketName string) error {
	_, err := s.S3.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Mgr) Bucket(bucketName string) types.BucketMgrInterface {
	return NewBucketMgr(s.S3, bucketName)
}

func (s *S3Mgr) BucketNative(bucketName string) *BucketMgr {
	return NewBucketMgr(s.S3, bucketName)
}

func (s *S3Mgr) CreateBucket(bucketName string) (types.BucketMgrInterface, error) {
	resp, err := s.S3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, err
	}
	_ = resp
	mgr := NewBucketMgr(s.S3, bucketName)
	return mgr, nil
}

func (s *S3Mgr) ListBucket() (result types.ListBucketResp, err error) {
	resp, err := s.ListBucketNative()
	if err != nil {
		return result, err
	}
	for _, v := range resp.Buckets {
		result.Bucket = append(result.Bucket, types.Bucket{
			CreationDate: *v.CreationDate,
			Name:         *v.Name,
		})
	}
	return result, nil
}

func (s *S3Mgr) ListBucketNative() (*s3.ListBucketsOutput, error) {
	resp, err := s.S3.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type BucketMgr struct {
	bucketName string
	*s3.S3
	types.Op
}

func NewBucketMgr(s3cli *s3.S3, bucketName string) *BucketMgr {
	return &BucketMgr{
		bucketName,
		s3cli,
		types.Op{},
	}
}

func (b *BucketMgr) PutObject(objectKey string, reader io.Reader, options ...types.OpOption) error {
	// 在这一层封装进度条, 默认值
	b.Op.ApplyOpts(options)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	//defer cancel()
	ctx := context.TODO()
	//配置函数
	optFun := func(u *s3manager.Uploader) {
		u.PartSize = b.PartSize // 分块大小,默认当文件体积超过100M开始进行分块上传, 默认5mb
		// u.LeavePartsOnError = true //如果上传失败不要删除分片
		u.Concurrency = b.ThreadCount //并发数
	}

	//_, err := b.PutObjectWithContext(ctx, s3.PutObjectInput{
	//			Bucket:   aws.String(b.bucketName),
	//			Key:      aws.String(objectKey),
	//			Body:     b.setupProgress(reader),
	//			Metadata: b.Metadata,
	//})
	//_ = optFun
	resp, err := s3manager.NewUploaderWithClient(b.S3).UploadWithContext(ctx,
		&s3manager.UploadInput{
			Bucket:   aws.String(b.bucketName),
			Key:      aws.String(objectKey),
			Body:     b.setupProgress(reader),
			Metadata: b.Metadata,
		}, optFun)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			//传文件超时
			return fmt.Errorf("upload canceled due to timeout, %v\n", err)
		} else {
			return fmt.Errorf("failed to upload object,code:%s %v\n", aerr.Code(), err)
		}
	}
	_ = resp.Location
	return nil
}

func (b *BucketMgr) setupProgress(reader io.Reader) io.Reader {
	if b.ProgressFn == nil {
		return reader
	}
	var body io.Reader
	switch r := reader.(type) {
	case readerAtSeeker:
		body = NewCustomReader(r, b.ProgressFn)
	default:
		body = reader
	}
	return body
}

func (b *BucketMgr) GetObject(objectKey string, options ...types.OpOption) (io.ReadCloser, int64, error) {
	b.Op.ApplyOpts(options)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	resp, err := b.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: &b.bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return nil, 0, err
	}
	size := *resp.ContentLength
	return resp.Body, size, err
}

func (b *BucketMgr) DeleteObjects(objectKeys []string, options ...types.OpOption) (result types.DeleteObjectResp, err error) {
	b.Op.ApplyOpts(options)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	d := make([]*s3.ObjectIdentifier, 0)
	for _, v := range objectKeys {
		d = append(d, &s3.ObjectIdentifier{
			Key: aws.String(v),
		})
	}
	resp, err := b.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(b.bucketName),
		Delete: &s3.Delete{
			Objects: d,
		},
	})
	if err != nil {
		return result, err
	}
	for _, v := range resp.Deleted {
		result.Keys = append(result.Keys, *v.Key)
	}
	return
}

func (b *BucketMgr) GetObjectSignUrl(objectKey string, expire time.Duration, options ...types.OpOption) (result string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	b.Op.ApplyOpts(options)
	// 先判断key是否存在
	req, _ := b.HeadObjectRequest(&s3.HeadObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(objectKey),
	})
	req.SetContext(ctx)
	if err = req.Send(); err != nil {
		var awsErr awserr.RequestFailure
		if errors.As(err, &awsErr) && awsErr.StatusCode() == 404 {
			return "", errors.New("目标文件不存在")
		}
		return "", err
	}
	credValue, err := b.Config.Credentials.Get()
	if err != nil {
		return "", err
	}
	hostBase := req.HTTPRequest.URL.Host
	expireTime := time.Now().Add(expire).Unix()
	proto := "https" //https
	signtext := fmt.Sprintf("GET\n\n\n%d\n/%s/%s", expireTime, b.bucketName, objectKey)
	hash := hmac.New(sha1.New, []byte(credValue.SecretAccessKey))
	hash.Write([]byte(signtext))
	signature := url.QueryEscape(base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	signUrl := fmt.Sprintf("%s://%s/%s?AWSAccessKeyId=%s&Expires=%d&Signature=%s", proto,
		hostBase, objectKey, credValue.AccessKeyID, expireTime, signature)
	return signUrl, err
}

func (b *BucketMgr) ListObject(prefix string, options ...types.OpOption) (result types.ListObjectResp, err error) {
	//超时时间30秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	b.Op.ApplyOpts(options)
	var pMaxKeys *int64 = nil
	if b.MaxKeys > 0 {
		pMaxKeys = aws.Int64(b.MaxKeys)
	}
	var delimiter *string
	if b.Delimiter != "" {
		delimiter = aws.String(b.Delimiter)
	}
	out, err := b.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket:    aws.String(b.bucketName),
		Prefix:    aws.String(prefix),
		MaxKeys:   pMaxKeys,
		Delimiter: delimiter,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				err = fmt.Errorf("bucket不存在.")
			case request.CanceledErrorCode:
				err = fmt.Errorf("获取对象列表超时,%s", aerr.Error())
			default:
				err = fmt.Errorf(aerr.Error())
			}
		}
		return
	}
	for _, v := range out.Contents {
		result.Objects = append(result.Objects, types.Object{
			ETag:         *v.ETag,
			Key:          *v.Key,
			Size:         *v.Size,
			LastModified: *v.LastModified,
		})
	}
	return result, nil
}
