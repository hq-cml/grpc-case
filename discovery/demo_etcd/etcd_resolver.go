/**
 * 从 etcd 解析服务节点
 */
package demo_etcd

import (
	"context"
	"encoding/json"
	"fmt"
	etcdV3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	tickerTime = 10 * time.Second
)

type etcdResolver struct {
	etcdCli       *etcdV3.Client               // etcd 客户端
	etcdAddrs     []string                     // etcd 地址
	dialTimeout   time.Duration                // 连接 etcd 超时时间
	mr            map[string]resolver.Resolver // 记录所有创建的解析器，同一个host只创建一个解析器
	targetNodeSet map[string]*Node             // 需要解析的目标节点
	serviceNodes  map[string]map[string]*Node  // 解析到的服务节点， host:addr:*Node

	tnsMux sync.RWMutex
	snsMux sync.RWMutex
	cancel context.CancelFunc
	mrMux  sync.RWMutex
	once   sync.Once
}

// 获取当前解析到的服务节点
func (e *etcdResolver) getServiceNodes(name string) []*Node {
	e.snsMux.RLock()
	defer e.snsMux.RUnlock()
	nodes := make([]*Node, 0)

	for _, n := range e.serviceNodes[name] {
		nodes = append(nodes, n)

	}
	return nodes
}

// 设置解析到的服务节点
func (e *etcdResolver) setServiceNodes(name string, nodes ...*Node) {
	e.snsMux.Lock()
	defer e.snsMux.Unlock()
	ns := e.serviceNodes[name]
	if ns == nil {
		ns = make(map[string]*Node)
	}
	for i := range nodes {
		fmt.Printf("resolver node [%s:%s]\n", name, nodes[i].Addr)
		ns[nodes[i].Addr] = nodes[i]
	}
	e.serviceNodes[name] = ns
}

// 溢出服务节点
func (e *etcdResolver) removeServiceNode(name, addr string) {
	e.snsMux.Lock()
	defer e.snsMux.Unlock()
	nodes := e.serviceNodes[name]
	if nodes == nil {
		return
	}
	delete(nodes, addr)
}

// 设置解析器
func (e *etcdResolver) setManuResolver(host string, m resolver.Resolver) {
	e.mrMux.Lock()
	defer e.mrMux.Unlock()
	e.mr[host] = m
}

// 根据host获取解析器
func (e *etcdResolver) getManuResolver(host string) (resolver.Resolver, bool) {
	e.mrMux.RLock()
	defer e.mrMux.RUnlock()
	if m, ok := e.mr[host]; ok {
		return m, ok
	}
	return nil, false
}

// 设置解析目标节点
func (e *etcdResolver) setTargetNode(host string) {
	e.tnsMux.Lock()
	e.targetNodeSet[host] = &Node{Name: host}
	e.tnsMux.Unlock()

	// 开始解析时进行相关操作，只执行一次
	e.once.Do(func() {
		var ctx context.Context
		ctx, e.cancel = context.WithCancel(context.Background())
		e.start(ctx)
	})
}

// 获取解析目标节点
func (e *etcdResolver) getTargetNodes() []*Node {
	e.tnsMux.RLock()
	defer e.tnsMux.RUnlock()

	nodes := make([]*Node, 0)
	for _, n := range e.targetNodeSet {
		nodes = append(nodes, n)
	}
	return nodes
}

// 解析所有需要解析的节点
func (e *etcdResolver) resolverAll(ctx context.Context) {
	nodes := e.getTargetNodes()
	for _, node := range nodes {
		// 根据前缀获取节点信息
		cctx, cancel := context.WithTimeout(context.Background(), e.dialTimeout)
		rsp, err := e.etcdCli.Get(cctx, node.buildPrefix(), etcdV3.WithPrefix())
		cancel()
		if err != nil {
			fmt.Printf("get service node [%s] error:%s\n", node.Name, err.Error())
			continue
		}
		for j := range rsp.Kvs {
			n := &Node{}
			err = json.Unmarshal(rsp.Kvs[j].Value, n)
			if err != nil {
				fmt.Printf("get service node [%s] error:%s\n", node.Name, err.Error())
				continue
			}
			e.setServiceNodes(node.Name, n)
		}
	}

	// 解析完服务节点后，更新到连接上
	e.mrMux.RLock()
	defer e.mrMux.RUnlock()
	for _, v := range e.mr {
		v.ResolveNow(resolver.ResolveNowOptions{})
	}
}

func (e *etcdResolver) start(ctx context.Context) {
	if len(e.etcdAddrs) == 0 {
		panic("demo_etcd should call SetDiscoveryAddress or set env DISCOVERY_HOST")
	}

	var err error
	e.etcdCli, err = etcdV3.New(etcdV3.Config{
		Endpoints:   e.etcdAddrs,
		DialTimeout: e.dialTimeout,
	})
	if err != nil {
		panic(err)
	}

	// 开始先全部解析
	e.resolverAll(ctx)

	ticker := time.NewTicker(tickerTime)

	// 定时解析
	go func() {
		for {
			select {
			case <-ticker.C:
				e.resolverAll(ctx)

			case <-ctx.Done():
				fmt.Println("resolver ticker exit")
				return
			}
		}
	}()

	// 每个节点watch变化
	nodes := e.getTargetNodes()
	for i := range nodes {
		go func(node *Node) {
			wc := e.etcdCli.Watch(ctx, node.buildPrefix(), etcdV3.WithPrefix())
			for {
				select {
				case rsp := <-wc:
					for _, event := range rsp.Events {
						switch event.Type {
						case etcdV3.EventTypePut:
							n := &Node{}
							err = json.Unmarshal(event.Kv.Value, n)
							if err != nil {
								fmt.Printf("unmarshal to node error:%s", err.Error())
								continue
							}
							e.setServiceNodes(node.Name, n)
						case etcdV3.EventTypeDelete:
							n := &Node{}
							err = json.Unmarshal(event.Kv.Value, n)
							if err != nil {
								fmt.Printf("unmarshal to node error:%s\n", err.Error())
								continue
							}
							e.removeServiceNode(node.Name, n.Addr)
						}
					}
				case <-ctx.Done():
					fmt.Println("resolver watcher exit")
					return
				}
			}
		}(nodes[i])
	}
}

func (e *etcdResolver) stop() {
	fmt.Println("resolver stop")
	e.cancel()
}

func etcdResolverInit() {
	envEtcdAddr := os.Getenv("DISCOVERY_HOST")
	eRegister = &etcdRegister{
		nodeSet:     make(map[string]*Node),
		etcdCli:     nil,
		dialTimeout: time.Second * 3,
		ttl:         3,
	}
	if len(envEtcdAddr) > 0 {
		eRegister.etcdAddrs = strings.Split(envEtcdAddr, ";")
	}
}
