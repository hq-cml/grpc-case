/**
 * 服务发现服务端，支持指定port
 * 在简单服务发现的例子中，服务端是比较简单的，只是简单的增加了一个port参数
 * 服务发现的文章，都在客户端实现（实际场景上，结合ETCD，则需要在服务端增加一些注册的逻辑）
 */
package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"grpc-case/common"
	"grpc-case/pb"
	"net"
	"sync"
)

var portStr *string = flag.String("p", common.BackEndPort0, "port")

// 业务自己的Server，实现各个服务端方法
type MyServer struct {
	pb.UnimplementedHelloServiceServer //首先包装一个UnimplementedServer，使自身成为HelloServiceServer接口的实现
}

// 实现业务代码
func (m *MyServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("Server[%v] Recv Request:%v\n", *portStr, req.Name)
	return &pb.HelloReply{
		Message: "Hello " + req.Name,
	}, nil
}

// 服务端提前启动2个实例（这里为了说明原理，所以手动启）
//const (
//	backend1 = "127.0.0.1:9090"
//	backend2 = "127.0.0.1:9091"
//)

// 服务启动起来
func main() {
	// 参数解析
	flag.Parse()

	// 创建监听端口
	listener, err := net.Listen("tcp", ":"+*portStr)
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
		fmt.Printf("Server Start! Port:%v\n", *portStr)
		defer wg.Done()
		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}
