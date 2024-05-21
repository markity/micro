package errcode

import (
	"fmt"

	"github.com/markity/micro/protocol"
)

type ErrCode struct {
	Code int
	Msg  string
	Data interface{}
}

func (errcode *ErrCode) Error() string {
	return fmt.Sprintf("ErrCode: code(%v), msg(%v), data(%v)", errcode.Code, errcode.Msg, errcode.Data)
}

type NetworkError struct{}

func (ne *NetworkError) Error() string {
	return "Network Error"
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
	case protocol.ProtocolErrorTypeParseProtoFailed:
		return "parse proto failed"
	case protocol.ProtocolErrorUnexpected:
		return "unexpected"
	}

	return "unknown"
}

func IsNetworkError(err error) bool {
	_, ok := err.(*NetworkError)
	return ok
}

func IsBizError(err error) bool {
	_, ok := err.(*ErrCode)
	return ok
}

func IsProtocolError(err error) bool {
	_, ok := err.(*ProtocolError)
	return ok
}
