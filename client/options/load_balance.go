package options

import (
	loadbalance "github.com/markity/micro/plugin/load_balance"
)

func WithLoadBalance(lb loadbalance.LoadBalance) Option {
	if lb == nil {
		panic("unexpected nil")
	}
	return Option{
		F: func(o *Options) {
			o.LoadBalance = lb
		},
	}
}
