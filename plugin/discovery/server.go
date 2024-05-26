package discovery

type Registery interface {
	// 在服务端运行成功后调用, 向注册中心注册, 如果error不为nil, 那么server.Run会原样返回error
	Register(serviceName string, addrPort string) error
	DeRegister(serviceName string, addrPort string)
}
