package command

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type fileObject struct {
	Key  string
	Path string
	File os.FileInfo
}

//遍历文件夹
func getFileObject(pathStr []string, r bool) (result []*fileObject, err error) {
	for _, p := range pathStr {
		//简化路径
		p, err = filepath.Abs(path.Clean(p))
		if err != nil {
			return
		}
		if strings.LastIndex(p, "*") != -1 {
			// 	path.Match() filepath.Glob()
			return nil, fmt.Errorf("Directory cannot appear '*'.\n")
		}
		file_info, err := os.Stat(p)
		if err != nil {
			return result, fmt.Errorf("os.Stat failed: %q, %v\n", p, err)
		}
		if file_info.IsDir() { //文件夹
			if !r {
				// 非递归出现了文件夹直接报错
				err = fmt.Errorf("%s is a directory", file_info.Name())
				return result, err
			}
			basePath := file_info.Name()
			err = filepath.Walk(p, func(pathStr string, f_info os.FileInfo, err error) error {
				if err != nil || f_info == nil {
					return err
				}
				if f_info.IsDir() {
					return nil
				}
				pathStr = strings.Replace(pathStr, "\\", "/", -1)
				f := &fileObject{File: f_info, Path: pathStr}
				pathStr = pathStr[len(p):]
				if strings.HasSuffix(p, "/") {
					pathStr = "/" + pathStr
				}
				f.Key = basePath + pathStr
				result = append(result, f)
				return nil
			})
		} else { //文件
			f := &fileObject{File: file_info, Key: file_info.Name(), Path: p}
			result = append(result, f)
		}
	}
	return
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
			} else {
				bucket = p[6 : 6+idx]
				objPrefix = p[7+idx:]

				// 判断后面没/且拓展类型不同的情况下才加/
				if !strings.HasSuffix(objPrefix, "/") {
					if len(files) > 1 || filepath.Ext(objPrefix) == "" {
						objPrefix += "/"
					} else if stat, err := os.Stat(files[0]); err == nil && stat.IsDir() {
						objPrefix += "/"
					}
				}
			}
			break
		} else {
			//最后一个必需是oss://
			if i == argsLen-1 {
				ExitWithError(ExitBadArgs, fmt.Errorf("parameter error. example: put FILE [FILE...] oss://BUCKET[/PREFIX].\n"))
			}
			//滤重
			for i--; i >= 0; i-- {
				if files == nil || files[i] == p {
					ExitWithError(ExitBadArgs, fmt.Errorf("dir or file repetition. %s\n", p))
				}
			}
			if p == "" {
				continue
			}
			files = append(files, strings.Replace(p, "\\", "/", -1))
		}
	}
	return
}

func ByteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
