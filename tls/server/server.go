/**
 * 简单服务端
 */
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"grpc-case/simple/pb"
	"net"
	"sync"
)

const (
	KeyPath = "/data/share/golang/src/github.com/hq-cml/grpc-case/tls/key/"
)

// 业务自己的Server，实现各个服务端方法
type MyServer struct {
	pb.UnimplementedHelloServiceServer //首先包装一个UnimplementedServer，使自身成为HelloServiceServer接口的实现
}

// 实现业务代码
func (m *MyServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Println("Recv Request:", req.Name)
	return &pb.HelloReply{
		Message: "Hello " + req.Name,
	}, nil
}

// 服务启动起来
func main() {
	// 载入证书：两个参数分别是自签证书 & 私钥
	creds, err := credentials.NewServerTLSFromFile(KeyPath+"test.pem", KeyPath+"test.key")
	if err != nil {
		panic(err)
	}

	// 创建监听端口
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	// 创建grpc服务，带上证书！！！
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	// 在grpc服务中，注册业务自己的服务
	pb.RegisterHelloServiceServer(grpcServer, &MyServer{})

	// 启动grpc服务
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		fmt.Println("TLS Server start!")
		defer wg.Done()
		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}
