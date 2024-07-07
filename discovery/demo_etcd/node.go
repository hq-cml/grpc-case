/*
*

	*服务节点信息

这里我们创建个 node 结构保存服务节点信息:
*/
package demo_etcd

import (
	"fmt"
	"strings"
)

type Node struct {
	Name string `json:"name"` // 名称
	Addr string `json:"addr"` // 地址
}

// 把服务名中的 . 转换为 /
func (s Node) transName() string {
	return strings.ReplaceAll(s.Name, ".", "/")
}

// 构建节点 key
func (s Node) buildKey() string {
	return fmt.Sprintf("/%s/%s", s.transName(), s.Addr)
}

// 构建节点前缀
func (s Node) buildPrefix() string {
	return fmt.Sprintf("/%s", s.transName())
}
