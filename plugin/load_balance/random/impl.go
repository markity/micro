package roundrobin

import (
	"math/rand"

	"github.com/markity/micro/plugin/discovery"
	loadbalance "github.com/markity/micro/plugin/load_balance"
)

type impl struct{}

func (i *impl) GetNext(ins []discovery.ServiceInstance) discovery.ServiceInstance {
	if len(ins) == 0 {
		panic("unexpected")
	}

	tobeRet := ins[rand.Int()%(len(ins)-1)]
	return tobeRet
}

func NewRoundRobin() loadbalance.LoadBalance {
	return &impl{}
}
