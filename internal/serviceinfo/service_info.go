package serviceinfo

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
