package protocol

import "github.com/markity/micro/errx"

// 为了在协议中封装这两种错误类型, 必须
type ErrXProtocol struct {
	BE  *errx.BizError
	SBE *errx.ServiceBusyError
}
