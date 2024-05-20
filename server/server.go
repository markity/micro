package server

import (
	goreactor "github.com/markity/go-reactor"
	eventloop "github.com/markity/go-reactor/pkg/event_loop"
	"github.com/markity/micro/handleinfo"
	"github.com/markity/micro/utils"
)

type MicroServer interface {
	Run() error
}

type microServer struct {
	baseLoop eventloop.EventLoop
	server   goreactor.TCPServer
}

func (ms *microServer) Run() error {
	return ms.Run()
}

var serviceNameContextKey = "svc_name"
var implementedServerContextKey = "implement"
var handlesContextKey = "handle_info"

func NewServer(serviceName string, addrPort string, implementedServer interface{}, handles map[string]handleinfo.HandleInfo) MicroServer {
	baseLoop := eventloop.NewEventLoop()
	reactorServer := goreactor.NewTCPServer(baseLoop, addrPort, utils.GetNThreads(), goreactor.RoundRobin())
	reactorServer.SetConnectionCallback(handleConn)
	reactorServer.SetMessageCallback(handleMessage)
	baseLoop.SetContext(serviceNameContextKey, serviceName)
	baseLoop.SetContext(implementedServerContextKey, implementedServer)
	baseLoop.SetContext(handlesContextKey, handles)
	return &microServer{
		baseLoop: baseLoop,
		server:   reactorServer,
	}
}
