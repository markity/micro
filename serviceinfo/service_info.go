package serviceinfo

type ServiceInfo interface {
	GetServiceName() string
	GetAllHandles() []HandleInfo
	GetHandle(name string) (HandleInfo, bool)
}

type HandleType int

const (
	// 两边都是proto消息的意思, request-response模式
	PROTO_TO_PROTO HandleType = iota
)

type HandleInfo struct {
	Name     string
	Type     HandleType
	Request  interface{}
	Response interface{}
}

type svcInfo struct {
	svcName   string
	svcHandle map[string]*HandleInfo
}

func (info *svcInfo) GetServiceName() string {
	return info.svcName
}

func (info *svcInfo) GetAllHandles() []HandleInfo {
	result := make([]HandleInfo, 0)
	for _, v := range info.svcHandle {
		result = append(result, *v)
	}
	return result
}

func (info *svcInfo) GetHandle(name string) (HandleInfo, bool) {
	h, ok := info.svcHandle[name]
	return *h, ok
}

func NewServiceInfo(serviceName string, handles []HandleInfo) ServiceInfo {
	hds := make(map[string]*HandleInfo)
	for _, v := range handles {
		cp := v
		hds[v.Name] = &cp
	}
	return &svcInfo{
		svcName:   serviceName,
		svcHandle: hds,
	}
}
