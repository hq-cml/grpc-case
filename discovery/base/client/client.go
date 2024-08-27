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
	"grpc-case/simple/pb"
	"time"
)

// 服务发现遵循RFC 3986，比如服务地址 xxx:///yyyy，则xxx就是Scheme，yyyy实际上是解析后的Path，把它作为service名字来使用
const (
	// Note: 注意有坑，值不能有大写字母！！
	myScheme      = "myscheme1"
	myServiceName = "myservicename1"
	address       = myScheme + ":///" + myServiceName // myScheme1://myServiceName1
)

// 服务端提前启动2个实例（这里为了说明原理，所以手动启，实际业务场景则是结合Etcd，自主发现）
const (
	backend1 = "127.0.0.1:9090"
	backend2 = "127.0.0.1:9091"
)

func main() {
	// 访问服务端address,创建连接conn,地址格式 myScheme:///myServiceName
	// 函数中会先根据myScheme这个scheme找到我们通过init函数注册的myBuilder，
	// 然后调用它的Build()方法构建我们自定义的myResolver，并调用ResolveNow()方法获取到服务端地址
	conn, err := grpc.NewClient(address,
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
	time.Sleep(2 * time.Second)
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
