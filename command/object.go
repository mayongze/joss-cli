package command

import (
	"container/list"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var (
	//上传超时时间单位秒
	uploadTmeout time.Duration
	//最小分片大小
	partSize int64
	//最大上传并发数
	concurrency int
)

func NewObjectPutCommand() *cobra.Command {
	bc := &cobra.Command{
		Use:   "put FILE [FILE...] oss://BUCKET[/PREFIX]",
		Short: "put object into the bucket.",

		Run: objectPutCommandFunc,
	}
	bc.Flags().DurationVar(&uploadTmeout, "upload-timeout", 0, "Upload timeout. unit:s")
	bc.Flags().Int64Var(&partSize,"part-size",50,"split upload size. unit:mb")
	bc.Flags().IntVarP(&concurrency,"concurrency","c",3,"oncurrent requests.")
	return bc
}

func objectPutCommandFunc(cmd *cobra.Command, args []string) {
	//解析参数
	bucket, objPrefix, files := getPutOp(args)
	s3Client,sess := NewS3Client(cmd)

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}
	//是否存在bucket
	_, err := s3Client.HeadBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				exitErrorf("bucket %s does not exist", bucket)
			}
		}
	}

	uploader := s3manager.NewUploader(sess)

	fileObjList,err := getFileObject(files)
	if err != nil {
		ExitWithError(ExitError,err)
	}
	for _,fileObj := range fileObjList{
		key := objPrefix + fileObj.Key
		err = writeFile(fileObj,bucket,key,uploader)
		if err != nil {ExitWithError(ExitError,err)}
		fmt.Println("文件写入成功: ", fileObj.Key)
	}

	//fmt.Println(bucket)
}

type fileObject struct {
	Key string
	Path string
	File os.FileInfo
}

//遍历文件夹
func getFileObject(path []string) (result []*fileObject,err error){
	for _,p := range path {
		//暂不支持windows下*号请求 */sss  *  /a/b/*  /a/b/sss*  aa/*b/dd b*   aa/ *b/../dd/ddf*ff   *b/../*/ddd   *b/../*/cc*
		//先进行路径处理把../case/../case/../command 处理成../command    ../case/./ 处理成../case  ../.././../
		// ../../../../case/../../../../../../   ../../../         .././../
		idx := strings.LastIndex(p,".")
		if idx > 3{
			arr := strings.Split(p,"/")
			l := list.New()
			for _,item := range arr {
				if l.Len() != 0{
					if item == "." {continue}
					data := l.Back()
					if item != ".." || data.Value.(string) == ".." || item == ""{
						l.PushBack(item)
						continue
					}
					l.Remove(data)
				}else{
					l.PushBack(item)
				}
			}
			// p   D:/aa/s/..   /exx/ee
			if p[0] == '/' {p = "/"} else { p=""}
			for e := l.Front(); e != nil; e = e.Next() {
				p = p + "/" + e.Value.(string)
			}
			p = p[1:]
		}

		if strings.LastIndex(p,"*") != -1 {
			return nil,fmt.Errorf("Directory cannot appear '*'.\n")
		}
		file_info, err := os.Stat(p)
		if err != nil {
			ExitWithError(ExitError,fmt.Errorf("os.Stat failed: %q, %v\n",p,err))
		}
			if file_info.IsDir(){ //文件夹
			basePath := file_info.Name()
			err = filepath.Walk(p,func(path string,f_info os.FileInfo,err error)error{
				if err != nil || f_info == nil {return err}
				if f_info.IsDir() {return nil}
				path = strings.Replace(path,"\\","/",-1)
				f := &fileObject{File:f_info,Path:p}
				path = path[len(p):]
				if strings.HasSuffix(p,"/") {
					path = "/" + path
				}
				if basePath[0] == '.' {
					f.Key = path[1:]
				}else{
					f.Key = basePath+path
				}
					result = append(result,f)
				return nil
			})
		}else{  //文件
			f := &fileObject{File:file_info,Key:file_info.Name(),Path:p}
			result = append(result,f)
		}

	}
	return
}

//写入文件
func writeFile(f_info *fileObject,bucket,key string,uploader *s3manager.Uploader) (err error) {
	size := f_info.File.Size()
	f, err  := os.Open(f_info.Path)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("failed to open file  %q, %v\n",f_info.Path,err)
	}
	//配合进度条的实现
	reader := &CustomReader{
		fp:   f,
		f_info:f_info,
		size: size,
	}
	 _, err = uploadWithSmr(reader,size,bucket,key,uploader)
	if err == nil {
		fmt.Fprintf(os.Stdout,"\n")
	}
	return err
}

type CustomReader struct {
	fp   *os.File
	f_info *fileObject
	size int64
	read int64
}
func (r *CustomReader) Read(p []byte) (int, error) {
	return r.fp.Read(p)
}

func (r *CustomReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := r.fp.ReadAt(p, off)
	if err != nil {
		return n, err
	}
	// Got the length have read( or means has uploaded), and you can construct your message
	atomic.AddInt64(&r.read, int64(n))
	// I have no idea why the read length need to be div 2,
	// maybe the request read once when Sign and actually send call ReadAt again
	// It works for me
	//log.Printf("Current upload file：%s",r.f_info.Key)

	fmt.Fprintf(os.Stdout,"\rTransfer Data, TotalBytes %d, ConsumedBytes: %d, progress:%d%%",
		r.size , r.read/2, int(float32(r.read*100/2)/float32(r.size)))
	return n, err
}

func (r *CustomReader) Seek(offset int64, whence int) (int64, error) {
	return r.fp.Seek(offset, whence)
}

//通过s3manager上传
func uploadWithSmr(reader io.Reader,size int64,bucket,key string,uploader *s3manager.Uploader) (*s3manager.UploadOutput,error) {
	//单位换成mb
	partLen := partSize * 1024 * 1024
	if size >= 5 * 1024 * 1024 * 1024 {
		return nil,fmt.Errorf("file maximum support 5gb. name=%s")
	}
	//最高支持10000个分片
	if size / 10000 > partLen {
		partLen = size / 100
	}
	input := &s3manager.UploadInput{
		Body:   reader,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	//获取一个根级上下文
	ctx := context.Background()
	var cancelFn func()
	if uploadTmeout > 0 {
		//设置一个timeout context
		ctx, cancelFn = context.WithTimeout(ctx, uploadTmeout)
		defer cancelFn()
	}
	//配置函数
	optFun := func(u *s3manager.Uploader) {
		u.PartSize = partLen // 分块大小,默认当文件体积超过100M开始进行分块上传
		//u.LeavePartsOnError = true //如果上传失败不要删除分片
		u.Concurrency = concurrency  //并发数
	}

	result, err := uploader.UploadWithContext(ctx,input,optFun)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			//传文件超时
			return nil,fmt.Errorf( "upload canceled due to timeout, %v\n", err)
		} else {
			return nil,fmt.Errorf("failed to upload object,code:%s %v\n", aerr.Code(),err)
		}
	}
	return result,err
}

func putObject(reader io.ReadSeeker,bucket,key string, s3cli *s3.S3) (*s3.PutObjectOutput, error){
	//获取一个根级上下文
	ctx := context.Background()
	var cancelFn func()
	if uploadTmeout > 0 {
		//设置一个timeout context
		ctx, cancelFn = context.WithTimeout(ctx, uploadTmeout)
		defer cancelFn()
	}
	result,err := s3cli.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:   reader,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return result,err
}

//获取put的参数
func getPutOp(args []string) (bucket, objPrefix string, files []string) {
	argsLen := len(args)
	if argsLen < 2 {
		ExitWithError(ExitBadArgs, fmt.Errorf("put command needs 2 argument. example: put FILE [FILE...] oss://BUCKET[/PREFIX].\n"))
	}
	for i, p := range args {
		if strings.HasPrefix(p, "oss://") {
			if i == 0 {
				ExitWithError(ExitBadArgs, fmt.Errorf("parameter error. example: put FILE [FILE...] oss://BUCKET[/PREFIX].\n"))
			}
			idx := strings.Index(p[6:], "/")
			if idx == -1 {
				bucket = p[6:]
				objPrefix = ""
			}else {
				bucket = p[6 : 6+idx]
				objPrefix = p[7+idx:]
				if !strings.HasSuffix(objPrefix,"/"){
					objPrefix += "/"
				}
			}
			break
		} else {
			//最后一个必需是oss://
			if i == argsLen-1 {
				ExitWithError(ExitBadArgs, fmt.Errorf("parameter error. example: put FILE [FILE...] oss://BUCKET[/PREFIX].\n"))
			}
			//滤重
			for i--;i >=0; i-- {
				if files[i] == p {
					ExitWithError(ExitBadArgs, fmt.Errorf("dir or file repetition. %s\n", p))
				}
			}
			if p == "" {continue}
			files = append(files, strings.Replace(p,"\\","/",-1))
		}
	}
	return
}
