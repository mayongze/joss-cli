## joss - 一款针对aws s3对象存储接口的命令行操作工具

使用golang编写的对象存储命令行操作工具.


## 安装

```
$ go get github.com/mayongze/joss-cli
$ cd $GOPATH/src/github.com/mayongze/joss-cli
$ make build
$ cd bin
```


## 使用方法

所有提供aws s3接口兼容的对象存储服务都可以使用joss.使用时需通过可选参数指定`accessKey` `secretKey` `edpoint`,也可使用`account`命令将ak sk信息保存在本地配置文件中.

目前兼容s3接口的对象存储提供商.joss默认使用京东云oss北京地区的endpoint.

- 京东云 https://oss-console.jdcloud.com 



## 操作命令

### account [options]

保存密钥命令,存储目录默认为`~/.jcloud`,在没有加参数的情况下默认输出当前本机当前存储使用的密钥.

可以通过环境变量`ACCESSKEY` `SECRETKEY`指定密钥参数.

*优先级：命令行参数>配置文件>环境变量*

##### Options

- --ak    需要保存的Accesskey
- --sk    需要保存的Secretkey

##### Examples

```bash
./joss account
# ak=************
# sk=************
./joss account --ak=******* --sk=*********
# OK
```



#### ls  \<oss://BUCKET[/PREFIX]>  [options]

显示对象信息.如没有指定前缀信息默认输出bucket列表

##### Options

- -l    以列的形式显示详细信息
- --max-keys    最大可输出的key数目

##### Examples

```bash
./joss ls
#bucket1 bucket2 bucket3
./joss ls oss://bucket1/joss
#file1 file2 file3
./joss ls oss://bucket1/joss -l
#bucket1 1234 2018-09-19 00:00 file1
#bucket1    4 2018-09-19 00:00 file2
#bucket1   34 2018-09-19 00:00 file3
```



#### put  \<FILE [FILE...] oss://BUCKET[/PREFIX]>  [options]

上传文件到对象存储.支持多个文件夹或者文件上传.

##### Options

- --upload-timeout    上传超时时间. 单位:秒（默认为0无限制）

- --concurrency    上传并发数  (默认为3)
- --part-size    单个分片上传的大小（默认为5MB）

##### Examples

```bash
./joss put /root/file1 /root/file2 oss://bucket1
#文件写入成功
./joss put /root/../export/* oss://bucket1
#文件写入成功
```

