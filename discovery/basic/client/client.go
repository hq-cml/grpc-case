/**
 * 服务发现客户端
 * 基础的服务发现逻辑，文章都在客户端
 */
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-case/common"
	"grpc-case/pb"
	"time"
)

// Note:
// 启动 go run client.go resolver.go
func main() {
	// 访问服务端address,创建连接conn,地址格式 myScheme:///myServiceName
	// 函数中会先根据myScheme这个scheme找到我们通过init函数注册的myBuilder，
	// （这里之前直接用common.BackEnd0，则会使用默认的Dns的Resolver）
	// 然后调用它的Build()方法构建我们自定义的myResolver，并调用ResolveNow()方法获取到服务端地址
	conn, err := grpc.NewClient(common.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig( // Note: 这里是在指定负载均衡的策略，如果不指定，则默认只会调用一个服务端实例，除非实例挂了才会切换
			fmt.Sprintf(`{"loadBalancingPolicy":"%v"}`, roundrobin.Name)),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 基于连接创建客户端,
	client := pb.NewHelloServiceClient(conn)

	fmt.Printf("Sleep!")
	time.Sleep(3 * time.Second)
	// 执行RPC调用
	for i := 0; i < 300; i++ {
		// 设置客户端访问超时时间1秒
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		rsp, err := client.SayHello(ctx, &pb.HelloRequest{
			Name: "msg_" + fmt.Sprintf("%v", i),
		})
		if err != nil {
			panic(err)
		}

		// 返回结果
		fmt.Println("Recv: ", rsp.Message)
		time.Sleep(1 * time.Second)
	}
}
