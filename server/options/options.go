package options

import "github.com/markity/micro/plugin/discovery"

type Options struct {
	Registry        discovery.Registery
	DoAfterRunHook  func()
	DoAfterStopHook func()
}

type Option struct {
	F func(*Options)
}
