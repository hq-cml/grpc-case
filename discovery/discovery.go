/**
 *
 */
package discovery

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

var eResolver *etcdResolver
var eRegister *etcdRegister

func init() {
	loggerInit()
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

func SetLogger(l *logrus.Logger) {
	logger = l
}
