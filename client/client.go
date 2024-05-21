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
	Call(handleName string, input proto.Message) (interface{}, error)
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
func (cli *microClient) Call(handleName string, input proto.Message) (interface{}, error) {
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
	var protocolError error
	hasProtoResult := false
	hasErrorResult := false
	var protoBytes []byte
	var errcodeBytes []byte
	go func() {
		defer func() {
			c <- struct{}{}
		}()

		// 建立连接
		beforeDial := time.Now()
		conn, err := net.DialTimeout("tcp", cli.serviceName, cli.timeoutDuration)
		if err != nil {
			networkError = &errcode.NetworkError{}
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
			networkError = &errcode.NetworkError{}
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
		case protocol.ProtocolErrorTypeParseProtoFailed:
			protocolError = &errcode.ProtocolError{
				Code: protocol.ProtocolErrorTypeParseProtoFailed,
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
			hasProtoResult = true
			bs := make([]byte, protoBodySize)
			_, err := io.ReadFull(conn, bs)
			if err != nil {
				networkError = &errcode.NetworkError{}
				return
			}
			protoBytes = bs
		}
		if errorCodeSize != 0 {
			hasErrorResult = true
			bs := make([]byte, errorCodeSize)
			_, err := io.ReadFull(conn, bs)
			if err != nil {
				networkError = &errcode.NetworkError{}
				return
			}
			errcodeBytes = bs
		}
	}()
	<-c
	if networkError != nil {
		return nil, networkError
	}
	if protocolError != nil {
		return nil, protocolError
	}

	var result1 interface{}
	var result2 error
	if hasProtoResult {
		protoVal := reflect.New(reflect.TypeOf(handle.Request).Elem()).Interface().(proto.Message)
		err := proto.Unmarshal(protoBytes, protoVal)
		if err != nil {
			return nil, &errcode.ProtocolError{
				Code: protocol.ProtocolErrorTypeParseProtoFailed,
			}
		}
		result1 = protoVal
	}
	if hasErrorResult {
		ec := errcode.ErrCode{}
		err := gob.NewDecoder(bytes.NewReader(errcodeBytes)).Decode(&ec)
		if err != nil {
			return nil, &errcode.ProtocolError{
				Code: protocol.ProtocolErrorTypeParseProtoFailed,
			}
		}
		result2 = &ec
	}

	return result1, result2
}
