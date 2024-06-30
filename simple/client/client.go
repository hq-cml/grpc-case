/**
 * 简单客户端
 */
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-case/simple/pb"
)

func main() {
	// 创建链接，此处禁用了安全传输，用了一个假的证书，insecure.NewCredentials()，没有加密和验证
	//conn, err := grpc.Dial("127.0.0.1: 9090", grpc.WithTransportCredentials(insecure.NewCredentials()))   # 废弃
	//conn, err := grpc.DialContext(context.Background(), "127.0.0.1: 9090", grpc.WithTransportCredentials(insecure.NewCredentials())) # 废弃
	conn, err := grpc.NewClient("127.0.0.1:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 建立连接并创建客户端
	client := pb.NewHelloServiceClient(conn)

	// 执行RPC调用
	rsp, err := client.SayHello(context.Background(), &pb.HelloRequest{
		Name: "far",
	})
	if err != nil {
		panic(err)
	}

	// 返回结果
	fmt.Println("Recv: ", rsp.Message)
}
