package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
	"reflect"
	"time"

	"github.com/markity/micro/errx"
	"github.com/markity/micro/handleinfo"
	"github.com/markity/micro/internal/protocol"
	"google.golang.org/protobuf/proto"
)

type MicroClient interface {
	Call(handleName string, input proto.Message) (interface{}, errx.ClientCallError)
}

type microClient struct {
	serviceName     string
	handles         map[string]handleinfo.HandleInfo
	timeoutDuration time.Duration
}

func NewClient(serviceName string, handles map[string]handleinfo.HandleInfo) MicroClient {
	return &microClient{
		serviceName:     serviceName,
		handles:         handles,
		timeoutDuration: time.Second * 3,
	}
}

// error is either errx.errx or errx.NetworkError or errx.ProtocolError
func (cli *microClient) Call(handleName string, input proto.Message) (interface{}, errx.ClientCallError) {
	inputMarshalBytes, err := proto.Marshal(input)
	if err != nil {
		panic(err)
	}

	handle, ok := cli.handles[handleName]
	if !ok {
		panic("unexpected")
	}

	c := make(chan struct{}, 1)
	var networkError error
	var protocolError bool
	var result1 interface{}
	var result2 errx.ClientCallError
	go func() {
		defer func() {
			c <- struct{}{}
		}()

		// 建立连接
		beforeDial := time.Now()
		conn, err := net.DialTimeout("tcp", cli.serviceName, cli.timeoutDuration)
		if err != nil {
			networkError = err
			return
		}
		defer conn.Close()
		ddl := beforeDial.Add(cli.timeoutDuration)
		conn.SetDeadline(ddl)
		var tmp [4]byte
		binary.BigEndian.PutUint32(tmp[:], uint32(len(handleName)))
		conn.Write(tmp[:])
		binary.BigEndian.PutUint32(tmp[:], uint32(len(inputMarshalBytes)))
		conn.Write(tmp[:])

		conn.Write([]byte(handleName))
		conn.Write(inputMarshalBytes)

		// 1字节指示错误, 4字节指示proto body长度, 4字节指示error长度
		var preBytes [9]byte
		_, err = io.ReadFull(conn, preBytes[:])
		if err != nil {
			networkError = err
			return
		}

		protoBodySize := binary.BigEndian.Uint32(preBytes[1:5])
		errorCodeSize := binary.BigEndian.Uint32(preBytes[5:9])

		code := protocol.ProtocolErrorType(preBytes[0])
		switch code {
		case protocol.ProtocolErrorTypeSuccess:
		default:
			protocolError = true
			return
		}
		// success的情况
		if protoBodySize != 0 {
			bs := make([]byte, protoBodySize)
			_, err := io.ReadFull(conn, bs)
			if err != nil {
				networkError = err
				return
			}

			protoVal := reflect.New(reflect.TypeOf(handle.Response).Elem()).Interface().(proto.Message)
			err = proto.Unmarshal(bs, protoVal)
			if err != nil {
				protocolError = true
				return
			}
			result1 = protoVal
		}
		if errorCodeSize != 0 {
			bs := make([]byte, errorCodeSize)
			_, err := io.ReadFull(conn, bs)
			if err != nil {
				networkError = err
				return
			}

			ec := protocol.ErrXProtocol{}
			err = gob.NewDecoder(bytes.NewReader(bs)).Decode(&ec)
			if err != nil {
				protocolError = true
				return
			}
			if ec.BE != nil {
				result2 = &clientCallError{
					isBizError: true,
					bizError:   ec.BE,
				}
			}
			if ec.SBE != nil {
				result2 = &clientCallError{
					isBusyError: true,
					busyError:   ec.SBE,
				}
			}
			if ec.BE != nil && ec.SBE != nil {
				protocolError = true
				return
			}
		}
	}()
	<-c
	if networkError != nil {
		return nil, &clientCallError{
			isNetworkError: true,
			networkError:   networkError,
		}
	}
	if protocolError {
		return nil, &clientCallError{
			isProtocolError: true,
		}
	}

	return result1, result2
}

type clientCallError struct {
	isNetworkError  bool
	isProtocolError bool
	isBusyError     bool
	isBizError      bool

	bizError     *errx.BizError
	networkError error
	busyError    *errx.ServiceBusyError
}

func (e *clientCallError) IsNetworkError() (error, bool) {
	return e.networkError, e.isNetworkError
}
func (e *clientCallError) IsBizError() (*errx.BizError, bool) {
	return e.bizError, e.isBizError
}
func (e *clientCallError) IsProtocolError() bool {
	return e.isProtocolError
}
func (e *clientCallError) IsBusyError() (*errx.ServiceBusyError, bool) {
	return e.busyError, e.isBusyError
}
