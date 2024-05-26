package options

import (
	"github.com/markity/micro/plugin/discovery"
	loadbalance "github.com/markity/micro/plugin/load_balance"
)

type Options struct {
	Discovery   discovery.Discovery
	LoadBalance loadbalance.LoadBalance
}

type Option struct {
	F func(*Options)
}
