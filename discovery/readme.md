# 服务发现的例子：

## 1. 基本原理
* 官方文档：https://github.com/grpc/grpc/blob/master/doc/naming.md
* 当服务端有多个实例的时候，就存在2个问题：服务发现和负载均衡
* gRPC中的默认服务发现是使用DNS，同时提供了一些接口，业务实现这些接口就可以自定义服务发现的功能
* 服务发现遵循RFC 3986，比如服务地址 xxx://yyyy，则xxx就是Scheme，yyyy为实际的service名字
* 负载均衡的算法也提前实现了一些基础的，可以直接选用，最常见的就是RoundRobin



## 核心概念
* 几个比较抽象的概念：
* Resolver是真正负责维护服务端地址列表的，用于将服务名解析成对应实例列表
* Builder是用来生成业务自己的Resolver


![avatar](img/name-resolver.png)


```
1）客户端启动时，注册自定义的 resolver 。
    一般在 init() 方法，构造自定义的 resolveBuilder，并将其注册到 grpc 内部的 resolveBuilder 表中（其实是一个全局 map，key 为协议名，value 为构造的 resolveBuilder）。
2）客户端启动时通过自定义 Dail() 方法构造 grpc.ClientConn 单例
    grpc.DialContext() 方法内部解析 URI，分析协议类型，并从 resolveBuilder 表中查找协议对应的 resolverBuilder。
    找到指定的 resolveBuilder 后，调用 resolveBuilder 的 Build() 方法，构建自定义 resolver，同时开启协程，通过此 resolver 更新被调服务实例列表。
    Dial() 方法接收主调服务名和被调服务名，并根据自定义的协议名，基于这两个参数构造服务的 URI
    Dial() 方法内部使用构造的 URI，调用 grpc.DialContext() 方法对指定服务进行拨号
3）grpc 底层 LB 库对每个实例均创建一个 subConnection，最终根据相应的 LB 策略，选择合适的 subConnection 处理某次 RPC 请求。
```
