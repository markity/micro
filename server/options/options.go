package options

import (
	"context"

	"github.com/markity/micro/plugin/discovery"
)

type Options struct {
	Registry        discovery.Registery
	DoAfterRunHook  func()
	DoAfterStopHook func()
	Ctx             context.Context
}

type Option struct {
	F func(*Options)
}
