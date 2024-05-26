package ipport

import "github.com/markity/micro/plugin/discovery"

type ipportDiscovery struct {
	ipport string
}

func (d *ipportDiscovery) GetInitInstances(serviceName string) []discovery.ServiceInstance {
	return []discovery.ServiceInstance{
		{ServiceName: serviceName, IPPort: d.ipport, Weight: 1},
	}
}

func (*ipportDiscovery) BlockGetNewInstances(serviceName string) []discovery.ServiceInstance {
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

func NewIPPortRegistery() discovery.Registery {
	return &ipportDiscovery{}
}
