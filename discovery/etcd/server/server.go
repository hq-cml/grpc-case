/**
 * 服务发现服务端，支持指定port
 * 在Etcd服务发现的例子中，服务端增加了注册Etcd的逻辑
 */
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	etcdCLientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"grpc-case/discovery/basic"
	"grpc-case/simple/pb"
	"net"
	"sync"
	"time"
)

var portStr *string = flag.String("p", "9090", "port")
var nameStr *string = flag.String("n", "myservicename1", "service name")
var scheme *string = flag.String("s", "etcd", "scheme")

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

	// 启动grpc服务
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		fmt.Printf("Server Start! Name:%v, Port:%v\n", *nameStr, *portStr)
		defer wg.Done()
		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())

	// 服务注册到Etcd
	go func() {
		for {
			err = registerToEtcd(ctx, *nameStr, addr)
			if err != nil && err.Error() == "over" {
				fmt.Println("退出")
				return
			} else {
				fmt.Println("遇到问题：" + err.Error() + ". 准备重新注册！！！！！！！")
			}
		}
	}()

	// 等待
	wg.Wait()

	// 关闭etcd相关操作
	cancel()
}

// 注册到Etcd
func registerToEtcd(ctx context.Context, service, addr string) error {
	// 创建客户端
	etcdCli, err := etcdCLientv3.New(etcdCLientv3.Config{
		Endpoints:   []string{basic.EtcdAddr},
		DialTimeout: basic.EtcdTimeout * time.Second,
	})
	if err != nil {
		return fmt.Errorf("etcdCLientv3.New Error:%v", err)
	}

	// 创建租约
	resp, err := etcdCli.Grant(ctx, 2) // ttl = 2，表示存活时间是2秒，如果不续约就消失
	if err != nil {
		return fmt.Errorf("Grant Err:%v", err)
	}
	fmt.Println("申请租约：Grant-------", resp.ID)

	// 注册
	key := basic.GenInstancePath(*scheme, service, addr)
	_, err = etcdCli.Put(ctx, key, addr, etcdCLientv3.WithLease(resp.ID))
	if err != nil {
		return fmt.Errorf("Etcd Put Err:%v", err)
	}
	fmt.Println("注册：Put-------", key, "=>", addr)

	// 租约保活
	respCh, err := etcdCli.KeepAlive(ctx, resp.ID)
	if err != nil {
		return fmt.Errorf("Etcd Keep Alive Err:%v", err)
	}

	// 异步监听保活的结果
	i := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("实例退出")
			return errors.New("over")
		case v, ok := <-respCh:
			if v == nil {
				// 如果失败了，要重新注册并续租
				fmt.Println("续租租约失效", ok)
				return errors.New("continue")
			} else {
				if i%100 == 0 {
					fmt.Println(time.Now().Format(time.DateTime),
						"续租租约成功"+fmt.Sprintf("(%v)", i), v, ok)
				}
				i++
			}
		}
	}

	// 测试租约回收
	//go func() {
	//	time.Sleep(10 * time.Second)
	//	_, err = etcdCli.Revoke(ctx, resp.ID) // 撤销租约
	//	if err != nil {
	//		panic(err)
	//	}
	//}()
}
