package options

import "github.com/markity/micro/plugin/discovery"

func WithDiscovery(discovery discovery.Discovery) Option {
	if discovery == nil {
		panic("unexpected nil")
	}
	return Option{
		F: func(o *Options) {
			o.Discovery = discovery
		},
	}
}
