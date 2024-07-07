/**
 * 基于Token认证的服务端
 */
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

// 模拟Token校验（这个在实际工程上放在拦截器里更合适）
func (m *MyServer) check(ctx context.Context) (bool, string) {
	// 从metadata中取出appId和appKey
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false, "获取失败"
	}
	s, _ := json.Marshal(md)
	fmt.Println("X---------", string(s))

	// 从上下文中取出client带过来的appId和appKey
	// 注意这里有个坑，必须全小写！metadata.FromIncomingContext有注释解释
	var appId, appKey string
	if v, ok := md["appid"]; ok {
		appId = v[0]
	}
	if v, ok := md["appkey"]; ok {
		appKey = v[0]
	}

	// 这里模拟从某个存储上，取出服务端维护的appId和appKey
	if appId != "123" && appKey != "abc" {
		return false, "校验失败"
	}
	return true, ""
}

// 实现业务代码
func (m *MyServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	// token校验
	ok, msg := m.check(ctx)
	if !ok {
		return nil, errors.New(msg)
	}

	fmt.Println("Recv Request:", req.Name)
	return &pb.HelloReply{
		Message: "Hello " + req.Name,
	}, nil
}

// 服务启动起来
func main() {
	// 载入证书：两个参数分别是自签证书 & 私钥
	//creds, err := credentials.NewServerTLSFromFile(KeyPath+"test.pem", KeyPath+"test.key")
	//if err != nil {
	//	panic(err)
	//}

	// 创建监听端口
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	// 创建grpc服务
	//grpcServer := grpc.NewServer(grpc.Creds(creds))
	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	// 在grpc服务中，注册业务自己的服务（也就是将自己的Server对象与grpc服务绑定）
	pb.RegisterHelloServiceServer(grpcServer, &MyServer{})

	// 启动grpc服务
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		fmt.Println("Token Server start!")
		defer wg.Done()
		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}
