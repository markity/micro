package discovery

type Discovery interface {
	GetInitInstances() []ServiceInstance     // NewClient时第一次调用, 会阻塞等待拿到初始化的instances
	BlockGetNewInstances() []ServiceInstance // 随后, 客户端会开启一个协程阻塞在此处, 获取更新的示例
}
