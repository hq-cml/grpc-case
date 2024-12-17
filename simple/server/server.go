/**
 * 简单服务端
 */
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"grpc-case/common"
	"grpc-case/pb"
	"net"
	"sync"
)

// 业务自己的Server，实现各个服务端方法
type MyServer struct {
	pb.UnimplementedHelloServiceServer //首先包装一个UnimplementedServer，使自身成为HelloServiceServer接口的实现
}

// 实现业务代码
func (m *MyServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Println("Recv Request:", req.Name)
	return &pb.HelloReply{
		Message: "Hello bar",
	}, nil
}

// 服务启动起来
func main() {
	// 创建监听端口
	listener, err := net.Listen("tcp", ":"+common.BackEndPort0)
	if err != nil {
		panic(err)
	}

	// 创建grpc服务
	grpcServer := grpc.NewServer()

	// 在grpc服务中，注册业务自己的服务（也就是将自己的Server对象与grpc服务绑定）
	pb.RegisterHelloServiceServer(grpcServer, &MyServer{})

	// 启动grpc服务
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		fmt.Println("Simple Server start!. Port:" + common.BackEndPort0)
		defer wg.Done()
		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}
