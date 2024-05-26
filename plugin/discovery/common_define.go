package discovery

type ServiceInstance struct {
	ServiceName string `json:"service_name"`
	IPPort      string `json:"ip_port"`
	Weight      int    `json:"weight"` // 权重, 当有多个选择时, 客户端会按权重进行选择
}
