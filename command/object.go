package command

import (
	"fmt"
	"github.com/mayongze/joss-cli/pkg/joss"
	"github.com/mayongze/joss-cli/pkg/joss/types"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var (
	//上传超时时间单位秒
	uploadTimeoutFlag string
	// uploadTimeout	time.Duration
	//最小分片大小
	partSize int64
	//最大上传并发数
	concurrency int
)

func init() {
	RootCmd.AddCommand(
		// ls oss://<bucket>/prefix
		NewObjectListCommand(),
		// put file file... oss://<bucket>/prefix
		NewObjectPutCommand(),

		NewSignUrlCommand(),
	)
}

//获取bucket 内的文件
func NewObjectListCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "ls oss://BUCKET[/PREFIX]",
		Aliases: []string{"list", "ll"},
		Short:   "获取oss里的列表信息,如需获取对象信息需要使用oss://指定bucket",

		Run: ossListCommandFunc,
	}
	c.Flags().BoolP("long", "l", false, "显示详细信息.")
	c.Flags().Int64("max-keys", 0, "keys的最大输出数目.")
	return c
}

func NewObjectPutCommand() *cobra.Command {
	bc := &cobra.Command{
		Use:   "put FILE [FILE...] oss://BUCKET[/PREFIX]",
		Short: "put object into the bucket.",

		Run: objectPutCommandFunc,
	}
	bc.Flags().StringVar(&uploadTimeoutFlag, "upload-timeout", "5s", "Upload timeout.")
	bc.Flags().Int64Var(&partSize, "part-size", 50, "split upload size. unit:mb")
	bc.Flags().IntVarP(&concurrency, "concurrency", "c", 3, "oncurrent requests.")

	bc.Flags().BoolVarP(&Force, "force", "f", false, "force updates")
	bc.Flags().BoolVarP(&Recursive, "recursive", "r", false, "Recursive upload, download or removal.")
	return bc
}

func NewSignUrlCommand() *cobra.Command {
	bc := &cobra.Command{
		Use:   "signurl oss://BUCKET/OBJECT <expiry>(10m|1h|1d)",
		Short: "Sign an OSS URL to provide limited public access with expiry",

		Run: signUrlCommandFunc,
	}

	return bc
}

func signUrlCommandFunc(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		ExitWithError(ExitBadArgs, fmt.Errorf("signurl command needs 2 argument. example: signurl oss://BUCKET/OBJECT <expiry>(10m|1h|1d).\n"))
	}
	param := args[0]
	bucket := ""
	key := ""
	if strings.HasPrefix(param, "oss://") {
		str := strings.SplitN(strings.TrimLeft(param, "oss://"), "/", 2)
		if len(str) == 1 {
			bucket = str[0]
		} else if len(str) == 2 {
			bucket = str[0]
			key = str[1]
		}
	} else {
		ExitWithError(ExitError, fmt.Errorf("bucket格式错误. Format: oss://bucket[/objectKey]"))
	}
	if bucket == "" {
		ExitWithError(ExitError, fmt.Errorf("bucket不能为空. Format: oss://bucket[/objectKey]"))
	}
	expire, err := time.ParseDuration(args[1])
	if err != nil {
		ExitWithError(ExitError, fmt.Errorf("expiry error, example: signurl oss://BUCKET/OBJECT <expiry>(10m|1h|1d)"))
	}
	url, err := joss.New(Endpoint, AccessKey, SecretKey, JossType).
		Bucket(bucket).GetObjectSignUrl(key, expire)
	if err != nil {
		ExitWithError(ExitError, err)
	}
	fmt.Println(url)
}

//获取对象存储列表
func ossListCommandFunc(cmd *cobra.Command, args []string) {
	//无附加参数获取bucket 列表
	if len(args) == 0 {
		bucketListCommandFunc(cmd, args)
		return
	}
	//解析参数  oss://bucket/aae/f
	param := args[0]
	bucket := ""
	prefix := ""
	if strings.HasPrefix(param, "oss://") {
		str := strings.SplitN(strings.TrimLeft(param, "oss://"), "/", 2)
		if len(str) == 1 {
			bucket = str[0]
		} else if len(str) == 2 {
			bucket = str[0]
			prefix = str[1]
		}
	} else {
		ExitWithError(ExitError, fmt.Errorf("bucket格式错误. Format: ls oss://bucket[/prefix]"))
	}
	if bucket == "" {
		ExitWithError(ExitError, fmt.Errorf("bucket不能为空. Format: ls oss://bucket[/prefix]"))
	}
	maxKeys, _ := cmd.Flags().GetInt64("max-keys")
	resp, err := joss.New(Endpoint, AccessKey, SecretKey, JossType).
		Bucket(bucket).ListObject(prefix, types.WithMaxKeys(maxKeys))
	if err != nil {
		ExitWithError(ExitError, err)
	}

	longMode, _ := cmd.Flags().GetBool("long")
	// 全部按行输出
	longMode = true
	if longMode || cmd.CalledAs() == "ll" {
		fmt.Printf("total %d\n", len(resp.Objects))
		for _, obj := range resp.Objects {
			//|s|%s %3d %s|s|%8s|%s %-7s %s %-3s
			// b kb mb gb
			fmt.Printf("%9s\t%6s\t%s\n",
				ByteCountBinary(obj.Size),
				obj.LastModified.Format("2006-01-02 15:04"),
				obj.Key)
		}
	} else {
		for _, obj := range resp.Objects {
			fmt.Printf("%-6s\t", obj.Key)
		}
		fmt.Print("\n")
	}
}

func objectPutCommandFunc(cmd *cobra.Command, args []string) {
	// 参数验证, 分片大小默认50 sdk默认5
	if partSize < 5 {
		partSize = 5
	}
	//解析参数
	bucket, objPrefix, files := getPutOp(args)
	// 批量上传进度条咋搞
	jcli := joss.New(Endpoint, AccessKey, SecretKey, JossType)
	// 	// 文件递归 -r
	filesObj, err := getFileObject(files, Recursive)
	if err != nil {
		ExitWithError(ExitError, err)
	}
	// 判断key是否存在,只判断第一个
	if !Force {
		f := filesObj[0]
		p := ""
		if strings.HasSuffix(objPrefix, "/") {
			p = fmt.Sprint(objPrefix, f.Key)
		} else if Recursive {
			p = fmt.Sprint(objPrefix, "/", f.Key)
		}
		resp, _ := jcli.Bucket(bucket).ListObject(p)
		if len(resp.Objects) != 0 || len(resp.Prefix) != 0 {
			ExitWithError(ExitError, fmt.Errorf("The target file already exists, try using -f flag."))
		}
	}

	pSuffix := fmt.Sprint(bucket, "/", objPrefix)
	for i, v := range filesObj {
		var p = ""
		// 有/则加文件名 或者没/但是是有递归参数
		if strings.HasSuffix(pSuffix, "/") {
			p = fmt.Sprint(pSuffix, v.Key)
		} else if Recursive {
			p = fmt.Sprint(pSuffix, "/", v.Key)
		}
		fmt.Printf("upload: '%s' -> 'oss://%s'  [%d of %d]\n", v.Key, p, i+1, len(filesObj))
		// 处理进度条
		start := time.Now()
		t := start
		var preConsumed int64 = 0
		var done = ""
		// p <- josstest/aaaz/
		if err = jcli.PutObjectFromFile(p, v.Path, types.WithProgress(func(totalBytes int64, consumedBytes int64) {
			if consumedBytes >= totalBytes {
				done = "  done\n"
			}
			now := time.Now()
			interval := now.Sub(t).Seconds()
			t = time.Now()

			diff := float64(consumedBytes - preConsumed)
			speed := int64(diff / interval)
			speedHuman := ByteCountBinary(speed)
			preConsumed = consumedBytes
			// 589588 of 589588   100% in    0s     2.01 MB/s  done
			fmt.Printf("\r%13d of %d\t%3.2f%% in\t%s\t%9s/s%s",
				consumedBytes, totalBytes, float32(consumedBytes*100)/float32(totalBytes),
				((now.Sub(start) / time.Second) * time.Second).String(), speedHuman, done)
		}), types.WithPartSize(partSize*1024*1024)); err != nil {
			ExitWithError(ExitError, err)
		}
	}
	// fmt.Println("文件全部写入成功: ", files)
}
