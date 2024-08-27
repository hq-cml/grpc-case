package base

import "strings"

const (
	EtcdAddr    = "127.0.0.1:2379"
	EtcdTimeout = 1
)

func GenBasePath(scheme, serviceName string) string {
	return "/" + strings.Join([]string{scheme, serviceName}, "/")
}

func GenInstancePath(scheme, serviceName, addr string) string {
	return strings.Join([]string{GenBasePath(scheme, serviceName), addr}, "/")
}
