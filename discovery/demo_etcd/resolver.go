/**
 * 实现 Resolver 接口：
 */
package demo_etcd

import (
	"fmt"
	"google.golang.org/grpc/resolver"
)

type IResolver interface {
	getServiceNodes(host string) []*Node
	setTargetNode(host string)
	setManuResolver(host string, m resolver.Resolver)
}

type manuResolver struct {
	cc     resolver.ClientConn
	target resolver.Target
	r      IResolver
}

func (m manuResolver) ResolveNow(options resolver.ResolveNowOptions) {
	nodes := m.r.getServiceNodes(m.target.URL.Host)
	addresses := make([]resolver.Address, 0)
	for i := range nodes {
		addresses = append(addresses, resolver.Address{
			Addr: nodes[i].Addr,
		})
	}
	if err := m.cc.UpdateState(resolver.State{
		Addresses: addresses,
	}); err != nil {
		fmt.Printf("resolver update cc state error:%s\n", err.Error())
	}
}

func (manuResolver) Close() {

}
