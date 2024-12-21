/**
 * 服务发现的核心！
 * 因为建立连接时使用我们自定义的 Scheme，而不是默认的dns，所以需要有和这个自定义的Scheme相对应的Resolver来解析才行
 * Builder采用Builder模式在包初始化时创建并注册构造自定义Resolver实例。当客户端通过Dial方法对指定服务进行拨号时
 * grpc resolver 查找注册的 Builder 实例调用其 Build() 方法构建自定义 Resolver。
 *
 * 官方注释，更加精准：
 *  A ResolverBuilder is registered for a scheme (in this example, "example" is
 *  the scheme). When a ClientConn is created for this scheme, the
 *  ResolverBuilder will be picked to build a Resolver. Note that a new Resolver
 *  is built for each ClientConn. The Resolver will watch the updates for the
 *  target, and send updates to the ClientConn.
 */
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	etcdCLientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"grpc-case/discovery/basic"
	"strings"
	"sync"
	"time"
)

// 将自定义的Resolver注册到grpc中
// 注意这里不是直接注册Resolver，而是注册的Builder，而Builder负责创建Resolver
// Builder有一个Scheme方法，用来指定自身的Key（grpc内部有Map[scheme]=>Builder）
func init() {
	etcdCli, err := etcdCLientv3.New(etcdCLientv3.Config{
		Endpoints:   []string{basic.EtcdAddr},
		DialTimeout: basic.EtcdTimeout * time.Second,
	})
	if err != nil {
		panic(err)
	}

	resolver.Register(&etcdBuilder{
		client: etcdCli,
	})
}

// 业务自己的Builder，实现resolver.ResolverBuilder接口
// 这个Builder将会被注册到resolver包当中，它的作用是用来生成业务自己的Resolver
type etcdBuilder struct {
	client *clientv3.Client
}

// 用来被注册的时候，生成key
func (*etcdBuilder) Scheme() string {
	return myScheme
}

// 创建并返回业务自己的Resolver实例
func (eb *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s, _ := json.Marshal(target.URL)
	fmt.Printf("Call--------Build. Parsed Target.URL: %v\n", string(s))
	targetName := target.Endpoint()
	// 初始化先读取路径下的配置
	var initAddrs []string
	getResp, err := eb.client.Get(context.Background(), basic.GenBasePath(myScheme, targetName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	} else {
		for _, kv := range getResp.Kvs {
			fmt.Println("X----------------", string(kv.Key), "=>", string(kv.Value))
			initAddrs = append(initAddrs, string(kv.Value))
		}
	}

	// 这里为了简便，直接创建Resolver
	myResolver := &etcdResolver{
		target: target,
		cc:     cc,
		addrsMap: map[string][]string{
			targetName: initAddrs,
		},
	}
	// 强制触发一次更新
	myResolver.ResolveNow(resolver.ResolveNowOptions{})

	// 触发一个watcher goroutine，从etcd里面定时更新服务列表到addrsMap
	go func() {
		//cctx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
		//defer cancel()
		fmt.Println("Watch----", basic.GenBasePath(myScheme, targetName))
		rch := eb.client.Watch(context.Background(), basic.GenBasePath(myScheme, targetName), clientv3.WithPrefix(), clientv3.WithRev(getResp.Header.Revision))
		for n := range rch {
			var needRefresh bool
			for _, ev := range n.Events {
				fmt.Printf("receive etcd event %v\n", ev)
				switch ev.Type {
				case mvccpb.PUT:
					addr := string(ev.Kv.Value)
					myResolver.mu.Lock()
					if !exist(myResolver.addrsMap[targetName], addr) {
						myResolver.addrsMap[targetName] = append(myResolver.addrsMap[targetName], addr)
						needRefresh = true
						fmt.Println("新增地址：", addr)
					}
					myResolver.mu.Unlock()
				case mvccpb.DELETE:
					path := string(ev.Kv.Key) // 删除仅能得到key
					tmp := strings.Split(path, "/")
					if len(tmp) == 0 {
						panic("wrong path")
					}
					addr := tmp[len(tmp)-1]
					myResolver.mu.Lock()
					if exist(myResolver.addrsMap[targetName], addr) {
						myResolver.addrsMap[targetName] = remove(myResolver.addrsMap[targetName], addr)
						needRefresh = true
						fmt.Println("删除地址：", addr)
					}
					myResolver.mu.Unlock()
				}
			}
			// 触发地址更新
			if needRefresh {
				myResolver.ResolveNow(resolver.ResolveNowOptions{})
			}
		}
	}()

	return myResolver, nil
}

// 业务自己的Resolver，实现resolver.Resolver接口
// 这个Resolver是真正负责维护服务端地址列表的，用于将服务名解析成对应实例列表
type etcdResolver struct {
	target   resolver.Target
	cc       resolver.ClientConn
	addrsMap map[string][]string // serviceName => []backendIpPort
	mu       sync.RWMutex
}

// 触发解析的逻辑
// Note: 公司的线上框架，这个函数啥也不干，只通过异步的goroutine来调用r.cc.UpdateState
func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {
	fmt.Println("Call--------ResolveNow")
	// 直接从map中取出对应的addrList
	r.mu.RLock()
	addrStrs := r.addrsMap[r.target.Endpoint()] // 这里其实就是path
	if len(addrStrs) == 0 {
		return
	}

	instanceList := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		instanceList[i] = resolver.Address{Addr: s}
	}
	r.mu.RUnlock()

	// 更新连接状态信息，即把从路由表中查到的 addrs 更新到底层的 connection 中.
	r.cc.UpdateState(resolver.State{Addresses: instanceList})
}

func (*etcdResolver) Close() {}

func exist(l []string, addr string) bool {
	for _, v := range l {
		if v == addr {
			return true
		}
	}
	return false
}

func remove(s []string, addr string) []string {
	var ret []string
	for _, v := range s {
		if v != addr {
			ret = append(ret, v)
		}
	}
	return ret
}
