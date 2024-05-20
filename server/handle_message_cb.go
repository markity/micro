package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"reflect"

	goreactor "github.com/markity/go-reactor"
	"github.com/markity/go-reactor/pkg/buffer"
	"github.com/markity/micro/errcode"
	"github.com/markity/micro/handleinfo"
	"github.com/markity/micro/protocol"
	"google.golang.org/protobuf/proto"
)

var double0Uint64Byts = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func handleMessage(conn goreactor.TCPConnection, buf buffer.Buffer) {
	if buf.ReadableBytes() < 8 {
		return
	}

	// 检查是否够取
	handleNameSize := binary.BigEndian.Uint64(buf.Peek()[0:4])
	protoBodySize := binary.BigEndian.Uint64(buf.Peek()[5:8])
	if buf.ReadableBytes() < int(handleNameSize)+int(protoBodySize)+8 {
		return
	}

	handinfos := conn.GetEventLoop().
		MustGetContext(handlesContextKey).(map[string]*handleinfo.HandleInfo)
	implement := conn.GetEventLoop().
		MustGetContext(implementedServerContextKey)

	handleName := string(buf.Peek()[8 : 8+handleNameSize])
	protoBody := buf.Peek()[8+handleNameSize : 8+handleNameSize+protoBodySize]
	buf.Retrieve(8 + int(handleNameSize) + int(protoBodySize))

	handle, ok := handinfos[handleName]
	if !ok {
		conn.Send([]byte{byte(protocol.ProtocolErrorTypeHandleNameInvalid)})
		conn.Send(double0Uint64Byts)
		return
	}

	reqReflectValue := reflect.New(reflect.TypeOf(handle.Request))
	implementReflectValue := reflect.ValueOf(implement)

	if err := proto.Unmarshal(protoBody, reqReflectValue.Interface().(proto.Message)); err != nil {
		conn.Send([]byte{byte(protocol.ProtocolErrorTypeParseProtoFailed)})
		conn.Send(double0Uint64Byts)
		return
	}

	method_, ok := implementReflectValue.Type().MethodByName(handleName)
	if !ok {
		panic("unexpected")
	}
	methodType := method_.Type
	in := reqReflectValue.Convert(methodType.In(0))
	results := implementReflectValue.Call([]reflect.Value{reflect.ValueOf(context.Background()), in})
	if len(results) != 2 {
		panic("unexpected")
	}
	result1, result2 := results[0], results[1]
	var protoBytes []byte
	var errorBytes []byte
	if !result1.IsNil() {
		var err error
		resultMessage := result1.Interface().(proto.Message)
		protoBody, err = proto.Marshal(resultMessage)
		if err != nil {
			panic(err)
		}
	}
	if !result2.IsNil() {
		resultErrcode := result2.Interface().(*errcode.ErrCode)
		buf := bytes.NewBuffer(nil)
		err := gob.NewEncoder(buf).Encode(resultErrcode)
		if err != nil {
			panic(err)
		}
		errorBytes = buf.Bytes()
	}

	var tmp [4]byte
	binary.BigEndian.PutUint32(tmp[:], uint32(len(protoBytes)))
	conn.Send(tmp[:])
	binary.BigEndian.PutUint32(tmp[:], uint32(len(errorBytes)))
	conn.Send(tmp[:])

	conn.Send(errorBytes)
	conn.Send(protoBytes)
}