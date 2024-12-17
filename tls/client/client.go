/**
 * 基于Tls的客户端
 */
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"grpc-case/common"
	"grpc-case/pb"
)

const (
	KeyPath = "/data/share/golang/src/github.com/hq-cml/grpc-case/tls/key/"
)

func main() {
	// 载入证书，同时需要指定域名访问，如果域名错误，也会失效
	// creds, err := credentials.NewClientTLSFromFile(KeyPath+"test.pem", "*.baidu.com") // 域名指定不正确，会得到错误
	creds, err := credentials.NewClientTLSFromFile(KeyPath+"test.pem", "*.hq.com")
	if err != nil {
		panic(err)
	}

	// 创建链接，此处设置证书
	//conn, err := grpc.NewClient(common.BackEnd0, grpc.WithTransportCredentials(insecure.NewCredentials())) // 不指定证书，将会得到错误
	conn, err := grpc.NewClient(common.BackEnd0, grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 基于连接创建客户端
	client := pb.NewHelloServiceClient(conn)

	// 执行RPC调用
	rsp, err := client.SayHello(context.Background(), &pb.HelloRequest{
		Name: "foo",
	})
	if err != nil {
		panic(err)
	}

	// 返回结果
	fmt.Println("Recv: ", rsp.Message)
}
