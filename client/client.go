package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"

	"github.com/markity/micro/errcode"
	"github.com/markity/micro/handleinfo"
	"github.com/markity/micro/protocol"
	"google.golang.org/protobuf/proto"
)

type MicroClient interface {
	Call(handleName string, input proto.Message) (interface{}, errcode.ClientCallError)
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

// error is either errcode.Errcode or errcode.NetworkError or errcode.ProtocolError
func (cli *microClient) Call(handleName string, input proto.Message) (interface{}, errcode.ClientCallError) {
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
	var protocolError *errcode.ProtocolError
	var result1 interface{}
	var result2 errcode.ClientCallError
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
			fmt.Println(err)
			networkError = err
			return
		}

		protoBodySize := binary.BigEndian.Uint32(preBytes[1:5])
		errorCodeSize := binary.BigEndian.Uint32(preBytes[5:9])

		errCode := protocol.ProtocolErrorType(preBytes[0])
		switch errCode {
		case protocol.ProtocolErrorTypeHandleNameInvalid:
			protocolError = &errcode.ProtocolError{
				Code: protocol.ProtocolErrorTypeHandleNameInvalid,
			}
			if protoBodySize != 0 || errorCodeSize != 0 {
				protocolError = &errcode.ProtocolError{
					Code: protocol.ProtocolErrorUnexpected,
				}
			}
			return
		case protocol.ProtocolErrorTypeServerParseProtoFailed:
			protocolError = &errcode.ProtocolError{
				Code: protocol.ProtocolErrorTypeServerParseProtoFailed,
			}
			if protoBodySize != 0 || errorCodeSize != 0 {
				protocolError = &errcode.ProtocolError{
					Code: protocol.ProtocolErrorUnexpected,
				}
			}
			return
		default:
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
				protocolError = &errcode.ProtocolError{
					Code: protocol.ProtocolErrorTypeClientParseProtoFailed,
				}
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

			ec := errcode.ErrCode{}
			err = gob.NewDecoder(bytes.NewReader(bs)).Decode(&ec)
			if err != nil {
				protocolError = &errcode.ProtocolError{
					Code: protocol.ProtocolErrorTypeClientParseProtoFailed,
				}
			}
			result2 = &clientCallError{
				isBizError: true,
				bizError:   &ec,
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
	if protocolError != nil {
		return nil, &clientCallError{
			isProtocolError: true,
			protocolError:   protocolError,
		}
	}

	return result1, result2
}

type clientCallError struct {
	isNetworkError  bool
	isProtocolError bool
	isBizError      bool

	bizError      *errcode.ErrCode
	networkError  error
	protocolError *errcode.ProtocolError
}

func (e *clientCallError) IsNetworkError() (error, bool) {
	return e.networkError, e.isNetworkError
}
func (e *clientCallError) IsBizError() (*errcode.ErrCode, bool) {
	return e.bizError, e.isBizError
}
func (e *clientCallError) IsProtocolError() (*errcode.ProtocolError, bool) {
	return e.protocolError, e.isProtocolError
}
