package options

import "github.com/markity/micro/plugin/discovery"

func WithRegistry(reg discovery.Registery) Option {
	if reg == nil {
		panic("unexpected nil")
	}
	return Option{
		F: func(o *Options) {
			o.Registry = reg
		},
	}
}
