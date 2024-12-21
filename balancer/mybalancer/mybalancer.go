package mybalancer

/*
 * 参照round robin，自己实现一个定制化的balancer
 * 功能：不管多少个实例，第一个实例90%的概率，第二个实例10%的概率，其他实例没有机会！
 * 这里的设计模式和resolover类似，也有一个Builder和Picker
 * Builder负责生成Picker，而Picker负责真正的负载均衡策略的视线
 */
import (
	"fmt"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

// Name is the name of round_robin balancer.
const Name = "my_balancer"

// newBuilder creates a new balancer builder.
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &myPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type myPickerBuilder struct{}

// Build方法会在连接状态改变时被调用, 永远传递最新的所有健康连接, 实现方需要通过这些信息实现出
func (*myPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	fmt.Printf("myPickerBuilder Picker: Build called with info: %v\n", info)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	// 去除最新鲜的健康地址
	scs := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for sc := range info.ReadySCs {
		scs = append(scs, sc)
	}

	// 将参数传递进来的最新健康连接保存
	return &myPicker{
		subConns: scs,
	}
}

type myPicker struct {
	// subConns is the snapshot of the balancer when this picker was
	// created. The slice is immutable. Each Get() will do a customized
	// selection from it and return the selected SubConn.
	subConns []balancer.SubConn
}

// 真正的负载均衡的策略，也就根据定制策略选择一个实例
func (p *myPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	if len(p.subConns) <= 1 {
		return balancer.PickResult{SubConn: p.subConns[0]}, nil
	}
	// 不管多少个实例，第一个实例90%的概率，第二个实例10%的概率，其他实例没有机会
	if 0 != rand.Intn(10) {
		return balancer.PickResult{SubConn: p.subConns[1]}, nil
	} else {
		return balancer.PickResult{SubConn: p.subConns[0]}, nil
	}
}
