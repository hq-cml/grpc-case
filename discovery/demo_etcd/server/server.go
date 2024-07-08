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
	etcdCLientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"grpc-case/simple/pb"
	"net"
	"strings"
	"sync"
	"time"
)

var portStr *string = flag.String("p", "9090", "port")

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

// 服务启动起来
func main() {
	// 参数解析
	flag.Parse()

	// 创建监听端口
	addr := "127.0.0.1:" + *portStr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	// 创建grpc服务
	grpcServer := grpc.NewServer()

	// 在grpc服务中，注册业务自己的服务（也就是将自己的Server对象与grpc服务绑定）
	pb.RegisterHelloServiceServer(grpcServer, &MyServer{})
	etcdCli, err := etcdCLientv3.New(etcdCLientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
		//DialTimeout: e.dialTimeout, TODO
	})
	if err != nil {
		panic(err)
	}
	err = Register(context.Background(), etcdCli, "A", addr)
	if err != nil {
		panic(err)
	}
	fmt.Println("X----------------------")
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

func Register(ctx context.Context, client *etcdCLientv3.Client, service, addr string) error {
	resp, err := client.Grant(ctx, 1)
	if err != nil {
		return fmt.Errorf("Grant Err:%v", err)
	}
	fmt.Println("Grant-------", resp.ID)

	key := strings.Join([]string{service, addr}, "/")
	_, err = client.Put(ctx, key, addr, etcdCLientv3.WithLease(resp.ID))
	if err != nil {
		return fmt.Errorf("Etcd Put Err:%v", err)
	}
	fmt.Println("Put-------", key, "=>", addr)
	// respCh 需要消耗, 不然会有 warning
	respCh, err := client.KeepAlive(ctx, resp.ID)
	if err != nil {
		return fmt.Errorf("Etcd Keep Alive Err:%v", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				//return nil
			case v, ok := <-respCh:
				if v == nil {
					fmt.Println("租约失效", ok)
				} else {
					fmt.Println(time.Now().Format(time.DateTime), "租约成功", v, ok)
				}
			}
		}
	}()

	// 测试租约回收
	//go func() {
	//	time.Sleep(10 * time.Second)
	//	_, err = client.Revoke(ctx, resp.ID) // 撤销租约
	//	if err != nil {
	//		panic(err)
	//	}
	//}()

	return nil
}
