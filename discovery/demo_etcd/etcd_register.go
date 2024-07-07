/**
 * 注册节点信息到 etcd
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
	DefaultEtcdAddr = "127.0.0.1:2379"
)

type etcdRegister struct {
	etcdCli     *etcdV3.Client // etcd句柄
	etcdAddrs   []string       // etcd服务地址
	dialTimeout time.Duration  // 连接超时时间

	nodeSet     map[string]*Node   // 注册节点set
	etcdLeaseId etcdV3.LeaseID     // etcd租约id
	ttl         int64              // 注册节点过期时间
	cancel      context.CancelFunc // 取消函数，用去结束注册任务
	once        sync.Once
}

func init() {
	envEtcdAddr := os.Getenv("DISCOVERY_HOST")
	if envEtcdAddr == "" {
		fmt.Printf("Use Default EtcdAddr :%v\n", DefaultEtcdAddr)
		envEtcdAddr = envEtcdAddr
	}
	eResolver = &etcdResolver{
		mr:            make(map[string]resolver.Resolver),
		dialTimeout:   time.Second * 3,
		targetNodeSet: make(map[string]*Node),
		serviceNodes:  make(map[string]map[string]*Node),
	}
	if len(envEtcdAddr) > 0 {
		eResolver.etcdAddrs = strings.Split(envEtcdAddr, ";")
	}

	var ctx context.Context
	ctx, eResolver.cancel = context.WithCancel(context.Background())
	eResolver.start(ctx)
}

// 新增注册的服务节点
func (e *etcdRegister) addServiceNode(node *Node) {
	e.nodeSet[node.buildKey()] = node

	// 新增注册节点的时候，开始执行注册任务
	e.once.Do(
		func() {

		})
}

// 开始注册任务
func (e *etcdRegister) start(ctx context.Context) {
	if len(e.etcdAddrs) == 0 {
		panic("demo_etcd should call SetEtcdAddress or set env DISCOVERY_HOST")
	}

	// 连接etcd
	var err error
	e.etcdCli, err = etcdV3.New(etcdV3.Config{
		Endpoints:   e.etcdAddrs,
		DialTimeout: e.dialTimeout,
	})

	if err != nil {
		panic(err)
	}

	// 创建租约
	cctx, cancel := context.WithTimeout(ctx, e.dialTimeout)
	rsp, err := e.etcdCli.Grant(cctx, e.ttl)
	if err != nil {
		panic(err)
	}
	cancel()

	e.etcdLeaseId = rsp.ID

	// 保活
	kc, err := e.etcdCli.KeepAlive(ctx, rsp.ID)
	if err != nil {
		fmt.Printf("etcd keepalive error:%s\n", err.Error())
	}

	go func() {
		for {
			select {
			case kaRsp := <-kc:
				if kaRsp != nil {
					e.register(ctx)
				}
			case <-ctx.Done():
				fmt.Println("register exit")
				return
			}
		}
	}()
}

// 注册节点
func (e *etcdRegister) register(ctx context.Context) {
	// 遍历所有的服务节点进行注册
	for _, n := range e.nodeSet {
		value, err := json.Marshal(n)
		if err != nil {
			fmt.Printf("json marshal node:%s error:%s\n", n.Name, err.Error())
			continue
		}
		// 使用租约id注册
		cctx, cancel := context.WithTimeout(ctx, e.dialTimeout)
		_, err = e.etcdCli.Put(cctx, n.buildKey(), string(value), etcdV3.WithLease(e.etcdLeaseId))
		cancel()

		if err != nil {
			fmt.Printf("put %s:%s to etcd with lease id %d error:%s\n", n.buildKey(), string(value), e.etcdLeaseId, err.Error())
			continue
		}
		fmt.Printf("put %s:%s to etcd with lease id %d\n", n.buildKey(), string(value), e.etcdLeaseId)
	}
}

// 停止注册任务
func (e *etcdRegister) stop() {
	fmt.Println("register stop")
	// 退出注册任务
	e.cancel()

	// 清理注册信息
	for _, n := range e.nodeSet {
		value, err := json.Marshal(n)
		if err != nil {
			fmt.Printf("json marshal node:%s error:%s\n", n.Name, err.Error())
			continue
		}
		cctx, cancel := context.WithTimeout(context.Background(), e.dialTimeout)
		_, _ = e.etcdCli.Delete(cctx, n.buildKey())
		cancel()
		fmt.Printf("delete %s:%s from etcd\n", n.buildKey(), string(value))
	}
}
