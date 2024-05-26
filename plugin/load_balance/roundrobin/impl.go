package roundrobin

import (
	"github.com/markity/micro/plugin/discovery"
	loadbalance "github.com/markity/micro/plugin/load_balance"
)

type impl struct {
	idx int
}

func (i *impl) GetNext(ins []discovery.ServiceInstance) discovery.ServiceInstance {
	if len(ins) == 0 {
		panic("unexpected")
	}

	if i.idx >= len(ins) {
		i.idx = 0
	}

	tobeRet := ins[i.idx]
	i.idx++
	return tobeRet
}

func NewRoundRobin() loadbalance.LoadBalance {
	return &impl{idx: 0}
}
