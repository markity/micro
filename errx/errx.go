package errx

import (
	"fmt"
)

// 服务端返回得error类型, 二取其一, BizError或ServiceBusyError
type ErrX interface {
	ErrXMessage() string
}

type BizError struct {
	Code      int
	Msg       string
	ExtraData string // 自定义数据
}

func (be *BizError) ErrXMessage() string {
	return fmt.Sprintf("biz error: code(%v) msg(%v) data(%v)", be.Code, be.Msg, be.ExtraData)
}

type ServiceBusyError struct {
	Msg string
}

func (sbe *ServiceBusyError) ErrXMessage() string {
	return fmt.Sprintf("service busy error: msg(%v)", sbe.Msg)
}

type ClientCallError interface {
	IsNetworkError() (error, bool)          // 将会返回net包的error错误
	IsBizError() (*BizError, bool)          // 业务错误, 带业务错误代码和信息, 也可以携带自定义数据
	IsBusyError() (*ServiceBusyError, bool) // 代表业务正忙
	IsProtocolError() bool                  // 协议错误, 这个错误仅仅是用来提醒开发者检查"自身错误"的，可能是客户端和服务端用的生成的代码版本不一致
	IsNoInstanceError() bool                // 服务发现没有发现实例

	// 熔断策略, 触发条件: NetworkError, BusyError

	// 限流策略: 返回BusyError
}
