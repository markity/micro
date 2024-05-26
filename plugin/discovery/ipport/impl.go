package ipport

import "github.com/markity/micro/plugin/discovery"

type ipportDiscovery struct {
	ipport string
}

func (ipportDiscovery) GetInitInstances(serviceName string) []discovery.ServiceInstance {
	return []discovery.ServiceInstance{
		{ServiceName: serviceName, IPPort: serviceName, Weight: 1},
	}
}

func (ipportDiscovery) BlockGetNewInstances(serviceName string) []discovery.ServiceInstance {
	// block forever
	select {}
}

func (*ipportDiscovery) Register(serviceName string, addrPort string) error {
	return nil
}

func (*ipportDiscovery) DeRegister(serviceName string, addrPort string) {
	// just do nothing
}

func NewIPPortDiscovery(ipport string) discovery.Discovery {
	return &ipportDiscovery{
		ipport: ipport,
	}
}

func NewIPPortRegistery(ipport string) discovery.Registery {
	return &ipportDiscovery{
		ipport: ipport,
	}
}

type Registery interface {
	// 在服务端运行成功后调用, 向注册中心注册, 如果error不为nil, 那么server.Run会原样返回error
	Register(serviceName string, addrPort string) error
	DeRegister(serviceName string, addrPort string)
}
