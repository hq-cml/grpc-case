/**
 * 基于Token认证的客户端
 */
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-case/simple/pb"
)

const (
	KeyPath = "/data/share/golang/src/github.com/hq-cml/grpc-case/tls/key/"
)

// 自实现Token认证，实现credentials.PerRPCCredentials接口
type MyClientTokenAuth struct {
}

// 模拟从url中、或者其他什么东西中、例如配置之类的，取出appId和appKey
// 这东西要带给服务端，去做多租校验
func (m *MyClientTokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appId":  "123",
		"appKey": "abc",
	}, nil
}

// 如果返回true，则可以实现Token认证和TLS认证的叠加
func (m *MyClientTokenAuth) RequireTransportSecurity() bool {
	return false
}

func main() {
	// 载入证书，同时需要指定域名访问，如果域名错误，也会失效
	// creds, err := credentials.NewClientTLSFromFile(KeyPath+"test.pem", "*.baidu.com") // 域名指定不正确，会得到错误
	//creds, err := credentials.NewClientTLSFromFile(KeyPath+"test.pem", "*.hq.com")
	//if err != nil {
	//	panic(err)
	//}

	// 创建连接
	//conn, err := grpc.NewClient("127.0.0.1:9090", grpc.WithTransportCredentials(creds))
	conn, err := grpc.NewClient("127.0.0.1:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 不使用tls
		grpc.WithPerRPCCredentials(new(MyClientTokenAuth))) // 使用自实现的Token
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 建立连接并创建客户端
	client := pb.NewHelloServiceClient(conn)

	// 执行RPC调用
	rsp, err := client.SayHello(context.Background(), &pb.HelloRequest{
		Name: "haha",
	})
	if err != nil {
		panic(err)
	}

	// 返回结果
	fmt.Println("Recv: ", rsp.Message)
}
