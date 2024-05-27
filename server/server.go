package server

import (
	"context"

	goreactor "github.com/markity/go-reactor"
	eventloop "github.com/markity/go-reactor/pkg/event_loop"
	"github.com/markity/micro/handleinfo"
	"github.com/markity/micro/internal/utils"
	"github.com/markity/micro/plugin/discovery/ipport"
	"github.com/markity/micro/server/options"
)

var defaultOps = []options.Option{
	options.WithRegistry(ipport.NewIPPortRegistery()),
	options.WithCtx(context.Background()),
}

type MicroServer interface {
	Run() error
	With(opts ...options.Option)
	Stop()
}

type microServer struct {
	serverStarted bool

	serviceName string
	addrPort    string
	options     *options.Options
	baseLoop    eventloop.EventLoop
	server      goreactor.TCPServer
}

func (ms *microServer) Run() error {
	// 先bind address, 再注册服务发现
	if !ms.serverStarted {
		if err := ms.server.Start(); err != nil {
			return err
		}
		ms.serverStarted = true
	}

	// 服务发现插件
	if ms.options.Registry != nil {
		err := ms.options.Registry.Register(ms.serviceName, ms.addrPort)
		if err != nil {
			return err
		}

		if ms.options.DoAfterRunHook != nil {
			ms.baseLoop.DoOnLoop(func(el eventloop.EventLoop) {
				ms.options.DoAfterRunHook()
			})
		}

		ms.baseLoop.DoOnStop(func(el eventloop.EventLoop) {
			ms.options.Registry.DeRegister(ms.serviceName, ms.addrPort)
			if ms.options.DoAfterStopHook != nil {
				ms.options.DoAfterStopHook()
			}
		})
	}

	ms.baseLoop.Loop()
	return nil
}

func (ms *microServer) Stop() {
	ms.baseLoop.Stop()
}

func (ms *microServer) With(opts ...options.Option) {
	for _, v := range opts {
		v.F(ms.options)
	}
}

var serviceNameContextKey = "svc_name"
var implementedServerContextKey = "implement"
var handlesContextKey = "handle_info"
var ctxContextKey = "ctx"

func NewServer(serviceName string, addrPort string, implementedServer interface{}, handles map[string]handleinfo.HandleInfo, opts ...options.Option) MicroServer {
	options := &options.Options{}
	for _, opt := range defaultOps {
		opt.F(options)
	}
	for _, opt := range opts {
		opt.F(options)
	}
	baseLoop := eventloop.NewEventLoop()
	reactorServer := goreactor.NewTCPServer(baseLoop, addrPort, utils.GetNThreads(), goreactor.RoundRobin())
	reactorServer.SetConnectionCallback(handleConn)
	reactorServer.SetMessageCallback(handleMessage)
	baseLoop.SetContext(serviceNameContextKey, serviceName)
	baseLoop.SetContext(implementedServerContextKey, implementedServer)
	baseLoop.SetContext(handlesContextKey, handles)
	baseLoop.SetContext(ctxContextKey, options.Ctx)
	_, loops := reactorServer.GetAllLoops()
	for _, loop := range loops {
		loop.SetContext(serviceNameContextKey, serviceName)
		loop.SetContext(implementedServerContextKey, implementedServer)
		loop.SetContext(handlesContextKey, handles)
		loop.SetContext(ctxContextKey, options.Ctx)
	}
	return &microServer{
		serviceName: serviceName,
		addrPort:    addrPort,
		baseLoop:    baseLoop,
		server:      reactorServer,
		options:     options,
	}
}
