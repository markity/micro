package options

import "github.com/markity/micro/plugin/discovery"

type Options struct {
	Registry discovery.Registery
}

type Option struct {
	F func(*Options)
}
