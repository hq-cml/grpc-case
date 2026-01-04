package common

import "strings"

const (
	BackEndPort0 = "9090"
	BackEndPort1 = "9091"
	BackEndPort2 = "9092"
)

const (
	BackEnd0 = "127.0.0.1:9090"
	BackEnd1 = "127.0.0.1:9091"
	BackEnd2 = "127.0.0.1:9092"
)

const (
	AppId  = "123"
	AppKey = "abc"
)

const (
	EtcdAddr    = "127.0.0.1:2379"
	EtcdTimeout = 1
)

// 生成服务在Etcd中的注册路径：
// $scheme/$serviceName/$addr
func GenInstancePath(scheme, serviceName, addr string) string {
	return strings.Join([]string{GenBasePath(scheme, serviceName), addr}, "/")
}

func GenBasePath(scheme, serviceName string) string {
	return "/" + strings.Join([]string{scheme, serviceName}, "/")
}

// 服务发现遵循RFC 3986，比如服务地址 xxx:///yyyy，则xxx就是Scheme，yyyy实际上是解析后的Path，把它作为service名字来使用
const (
	// Note: 注意有坑，值不能有大写字母！！
	MyScheme      = "myscheme1"
	MyServiceName = "myservicename1"
	Address       = MyScheme + ":///" + MyServiceName
)

const (
	// Note: 注意有坑，值不能有大写字母！！
	MySchemeEtcd      = "etcd"
	MyServiceNameEtcd = "myservicename_etcd"
	AddressEtcd       = MySchemeEtcd + ":///" + MyServiceNameEtcd
)
