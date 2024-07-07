# 安装步骤：

## 1. 安装protocol buffers
* 简称Protobuf，是google开发的额一种协议。protoc就是pb的核心程序编译器，主要就是通过它编译proto文件，对结构数据进行序列化和反序列化。

```
安装方式：
    从github上下载二进制文件，https://github.com/protocolbuffers/protobuf/releases 
    然后解压添加到PATH路径即可。
```

***

## 2. 安装Protobuf编译器插件
* 负责生成除了序列化和反序列化之外的go相关代码，比如通信协议、程序框架等等。这里面有两个程序protoc-gen-go和protoc-gen-go-grpc。

```
安装方式：
    go get google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    
    go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    
    ps：
    此时这两个程序已经会出现在GOPATH/bin目录下，这个目录需要添加到$PATH中。
```

***

## 3. 安装GRPC
* grpc框架的核心库文件，grpc包的go版本库代码

```
安装方式：
    go get google.golang.org/grpc
```

# 目录：
```
simple：
    最简单直接的client和server的例子，直击问题本质
tls
    基于tls认证的server和client，使用自建证书，说明问题
token
    基于token认证的server和client（简单理解就是多租户服务的ak,sk)
discovery_base
    简单的服务发现例子，不涉及Etcd，说明问题本质原理
```