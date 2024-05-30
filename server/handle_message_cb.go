package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"reflect"

	goreactor "github.com/markity/go-reactor"
	"github.com/markity/go-reactor/pkg/buffer"
	"github.com/markity/micro/errx"
	"github.com/markity/micro/handleinfo"
	"github.com/markity/micro/internal/protocol"
	internal_protocol "github.com/markity/micro/internal/protocol"
	"github.com/markity/micro/internal/utils"
	"google.golang.org/protobuf/proto"
)

var double0Uint32Byts = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func handleMessage(conn goreactor.TCPConnection, buf buffer.Buffer) {
	qpsLimitEnabled := false
	windowLimitIface := conn.GetEventLoop().MustGetContext(windowLimitKey)
	var windowLimit *utils.WindowLimit
	if windowLimitIface != nil {
		qpsLimitEnabled = true
		windowLimit = windowLimitIface.(*utils.WindowLimit)
	}

	readableLength := buf.ReadableBytes()
	if readableLength < 8 {
		return
	}

	// 检查是否够取
	handleNameSize := binary.BigEndian.Uint32(buf.Peek()[0:4])
	protoBodySize := binary.BigEndian.Uint32(buf.Peek()[4:8])
	if readableLength < int(handleNameSize)+int(protoBodySize)+8 {
		return
	}

	handinfos := conn.GetEventLoop().
		MustGetContext(handlesContextKey).(map[string]handleinfo.HandleInfo)
	implement := conn.GetEventLoop().
		MustGetContext(implementedServerContextKey)

	handleName := string(buf.Peek()[8 : 8+handleNameSize])
	protoBody := buf.Peek()[8+handleNameSize : 8+handleNameSize+protoBodySize]
	buf.Retrieve(8 + int(handleNameSize) + int(protoBodySize))

	handle, ok := handinfos[handleName]
	if !ok {
		conn.Send([]byte{byte(protocol.ProtocolErrorUnexpected)})
		conn.Send(double0Uint32Byts)
		return
	}

	reqReflectValue := reflect.New(reflect.TypeOf(handle.Request).Elem())
	implementReflectValue := reflect.ValueOf(implement)

	if err := proto.Unmarshal(protoBody, reqReflectValue.Interface().(proto.Message)); err != nil {
		conn.Send([]byte{byte(protocol.ProtocolErrorUnexpected)})
		conn.Send(double0Uint32Byts)
		return
	}

	method_, ok := implementReflectValue.Type().MethodByName(handleName)
	if !ok {
		panic("unexpected")
	}
	methodType := method_.Type
	in := reqReflectValue.Convert(methodType.In(2))

	// 检查限流策略
	var results []reflect.Value
	if qpsLimitEnabled && !windowLimit.Add() {
		results = append(results, reflect.ValueOf(func() {}),
			reflect.ValueOf(&errx.ServiceBusyError{Msg: "service is busy(trigger qps limit)"}))
	} else {
		results = method_.Func.Call([]reflect.Value{implementReflectValue,
			reflect.ValueOf(conn.GetEventLoop().MustGetContext(ctxContextKey).(context.Context)), in})
	}
	if len(results) != 2 {
		panic("unexpected")
	}
	result1, result2 := results[0], results[1]
	var protoBytes []byte
	var errorBytes []byte
	if !result1.IsNil() {
		resultMessage := result1.Interface().(proto.Message)
		marshalBytes, err := proto.Marshal(resultMessage)
		if err != nil {
			panic(err)
		}
		protoBytes = marshalBytes
	}
	if !result2.IsNil() {
		var resultErrX internal_protocol.ErrXProtocol
		switch v := result2.Interface().(type) {
		case *errx.BizError:
			resultErrX.BE = v
		case *errx.ServiceBusyError:
			resultErrX.SBE = v
		default:
			panic("error must be one of *errx.BizError or *errx.ServiceBusyError")
		}
		buf := bytes.NewBuffer(nil)
		err := gob.NewEncoder(buf).Encode(resultErrX)
		if err != nil {
			panic(err)
		}
		errorBytes = buf.Bytes()
	}

	conn.Send([]byte{byte(protocol.ProtocolErrorTypeSuccess)})
	var tmp [4]byte
	binary.BigEndian.PutUint32(tmp[:], uint32(len(protoBytes)))
	conn.Send(tmp[:])
	binary.BigEndian.PutUint32(tmp[:], uint32(len(errorBytes)))
	conn.Send(tmp[:])

	conn.Send(protoBytes)
	conn.Send(errorBytes)
}
