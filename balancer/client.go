/**
 * 基于Etcd服务发现客户端
 * 同时附带一个自己实现的balancer
 */
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-case/balancer/mybalancer"
	"grpc-case/common"
	_ "grpc-case/discovery/etcd/client/resolver" // 这个很重要，注册基于etcd的resolver
	"grpc-case/pb"
	"time"
)

// 启动go run client.go
func main() {
	// 访问服务端address,创建连接conn,地址格式 myScheme:///myServiceName
	// 函数中会先根据myScheme这个scheme找到我们通过init函数注册的myBuilder，
	// 然后调用它的Build()方法构建我们自定义的myResolver，并调用ResolveNow()方法获取到服务端地址
	conn, err := grpc.NewClient(common.AddressEtcd,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig( // Note: 这里是在指定负载均衡的策略，如果不指定，则默认只会调用一个服务端实例，除非实例挂了才会切换
			//fmt.Sprintf(`{"loadBalancingPolicy":"%v"}`, roundrobin.Name)),
			fmt.Sprintf(`{"loadBalancingPolicy":"%v"}`, mybalancer.Name)),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 基于连接创建客户端,
	client := pb.NewHelloServiceClient(conn)

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
