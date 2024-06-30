#!/bin/bash

# 生成代码
protoc --go_out=. hello.proto           # 生成hello.pb.go，主要是一个结构的定义
protoc --go-grpc_out=. hello.proto      # 生成hello_grpc.pb.go，主要是程序的框架

# 参数：paths=source_relative，表示输出文件和输入文件位于同一个目录中
#           =import，表示输出文件将存在在以Go软件包导入路径（go_package）命名的目录中
# 暂时没有看出特别大的区别：protoc --go_out=. --go_opt=paths=source_relative hello.proto
