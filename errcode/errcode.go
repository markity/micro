package errcode

import (
	"fmt"

	"github.com/markity/micro/protocol"
)

// biz error define
type ErrCode struct {
	Code int
	Msg  string
	Data interface{}
}

func (errcode *ErrCode) Error() string {
	return fmt.Sprintf("ErrCode: code(%v), msg(%v), data(%v)", errcode.Code, errcode.Msg, errcode.Data)
}

type ClientCallError interface {
	IsNetworkError() (error, bool)
	IsBizError() (*ErrCode, bool)
	IsProtocolError() (*ProtocolError, bool)
}

type ProtocolError struct {
	Code protocol.ProtocolErrorType
}

func (pe *ProtocolError) Error() string {
	switch pe.Code {
	case protocol.ProtocolErrorTypeSuccess:
		return "success"
	case protocol.ProtocolErrorTypeHandleNameInvalid:
		return "handle name invalid"
	case protocol.ProtocolErrorTypeServerParseProtoFailed:
		return "server parse proto failed"
	case protocol.ProtocolErrorTypeClientParseProtoFailed:
		return "client parse proto failed"
	case protocol.ProtocolErrorUnexpected:
		return "unexpected"
	}

	return "unknown"
}
