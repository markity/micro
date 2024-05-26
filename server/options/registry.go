package options

import "github.com/markity/micro/plugin/discovery"

func WithRegistry(reg discovery.Registery) Option {
	return Option{
		F: func(o *Options) {
			o.Registry = reg
		},
	}
}
