/**
 *
 */
package demo_etcd

var eResolver *etcdResolver
var eRegister *etcdRegister

func init() {
	etcdRegisterInit()
	etcdResolverInit()
}

func SetDiscoveryAddress(address []string) {
	if len(address) == 0 {
		return
	}
	eResolver.etcdAddrs = address
	eRegister.etcdAddrs = address
}

func AddServiceNode(node *Node) {
	eRegister.addServiceNode(node)
}
