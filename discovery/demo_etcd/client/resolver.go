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
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

// 将自定义的Resolver注册到grpc中
// 注意这里不是直接注册Resolver，而是注册的Builder，而Builder负责创建Resolver
// Builder有一个Scheme方法，用来指定自身的Key（grpc内部有Map[scheme]=>Builder）
func init() {
	resolver.Register(&myBuilder{})
}

// 业务自己的Builder，实现resolver.ResolverBuilder接口
// 这个Builder将会被注册到resolver包当中，它的作用是用来生成业务自己的Resolver
type myBuilder struct {
	client *clientv3.Client
}

// 用来被注册的时候，生成key
func (*myBuilder) Scheme() string {
	return myScheme
}

// 创建并返回业务自己的Resolver实例
func (*myBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s, _ := json.Marshal(target.URL)
	fmt.Printf("Call--------Build. Parsed Target.URL: %v\n", string(s))

	// TODO 实际上，应该是根据不同的target，生成不用的Resolver

	// 这里为了简便，直接创建Resolver
	myResolver := &myResolver{
		target: target,
		cc:     cc,
		addrsMap: map[string][]string{
			myServiceName: []string{
				backend1,
				backend2,
			},
		},
	}

	// TODO 生成环境下，这里可以触发一个watcher goroutine，从etcd里面定时更新服务列表到addrsMap

	// 强制触发一次更新
	myResolver.ResolveNow(resolver.ResolveNowOptions{})
	return myResolver, nil
}

// 业务自己的Resolver，实现resolver.Resolver接口
// 这个Resolver是真正负责维护服务端地址列表的，用于将服务名解析成对应实例列表
type myResolver struct {
	target   resolver.Target
	cc       resolver.ClientConn
	addrsMap map[string][]string // serviceName => []backendIpPort
}

// 触发解析的逻辑
func (r *myResolver) ResolveNow(o resolver.ResolveNowOptions) {
	fmt.Println("Call--------ResolveNow")
	// 直接从map中取出对应的addrList
	addrStrs := r.addrsMap[r.target.Endpoint()]
	instanceList := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		instanceList[i] = resolver.Address{Addr: s}
	}

	// 更新连接状态信息，即把从路由表中查到的 addrs 更新到底层的 connection 中.
	r.cc.UpdateState(resolver.State{Addresses: instanceList})
}

func (*myResolver) Close() {}
