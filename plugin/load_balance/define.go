package loadbalance

import "github.com/markity/micro/plugin/discovery"

type LoadBalance interface {
	GetNext([]discovery.ServiceInstance) discovery.ServiceInstance
}
