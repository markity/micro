package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"math"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/markity/micro/client/options"
	"github.com/markity/micro/errx"
	"github.com/markity/micro/handleinfo"
	"github.com/markity/micro/internal/protocol"
	"github.com/markity/micro/plugin/discovery"
	"github.com/markity/micro/plugin/load_balance/roundrobin"
	"google.golang.org/protobuf/proto"
)

func timeMinor(a time.Time, b time.Time) time.Time {
	if a.UnixMilli() < b.UnixMilli() {
		return a
	}

	return b
}

var defaultOpts = []options.Option{
	options.WithLoadBalance(roundrobin.NewRoundRobin()),
	options.WithTimeout(time.Second * 3),
	options.WithRetry(options.Retry{
		MaxRetryTimes:    0,
		MaxTotalDuration: time.Second * 3,
		RetrySameNode:    false,
	}),
}

type MicroClient interface {
	Call(handleName string, input proto.Message) (interface{}, errx.ClientCallError)
}

type microClient struct {
	serviceName string
	handles     map[string]handleinfo.HandleInfo
	ops         *options.Options

	instanceMu sync.Mutex
	instances  []discovery.ServiceInstance
}

func NewClient(serviceName string, handles map[string]handleinfo.HandleInfo, opts ...options.Option) MicroClient {
	var ops options.Options
	for _, v := range defaultOpts {
		v.F(&ops)
	}
	for _, v := range opts {
		v.F(&ops)
	}
	if ops.Discovery == nil {
		panic("discovery is necessary")
	}
	var instances []discovery.ServiceInstance = ops.Discovery.GetInitInstances(serviceName)
	if ops.Discovery != nil {
		instances = ops.Discovery.GetInitInstances(serviceName)
	}
	cli := &microClient{
		serviceName: serviceName,
		handles:     handles,
		ops:         &ops,
		instances:   instances,
	}
	go func() {
		var mu = &cli.instanceMu
		var ins = &cli.instances
		var disc = ops.Discovery
		for {
			newInstances := disc.BlockGetNewInstances(serviceName)
			mu.Lock()
			*ins = newInstances
			mu.Unlock()
		}
	}()

	return cli
}

// error is either errx.errx or errx.NetworkError or errx.ProtocolError
func (cli *microClient) Call(handleName string, input proto.Message) (interface{}, errx.ClientCallError) {
	cli.instanceMu.Lock()
	if len(cli.instances) == 0 {
		cli.instanceMu.Unlock()
		return nil, &clientCallError{
			IsNoInstance: true,
		}
	}
	ins := cli.ops.LoadBalance.GetNext(cli.instances)
	cli.instanceMu.Unlock()

	beforeDial := time.Now()

	retryPolicy := cli.ops.RetryPolocy
	retryEnabled := false
	var retryTotalDeadline time.Time
	if retryPolicy.MaxRetryTimes != 0 {
		if retryPolicy.MaxRetryTimes < 0 {
			retryPolicy.MaxRetryTimes = math.MaxInt
		}
		retryEnabled = true
		retryTotalDeadline = beforeDial.Add(cli.ops.RetryPolocy.MaxTotalDuration)
	}

	inputMarshalBytes, err := proto.Marshal(input)
	if err != nil {
		panic(err)
	}

	handle, ok := cli.handles[handleName]
	if !ok {
		panic("unexpected")
	}

	ins = cli.ops.LoadBalance.GetNext(cli.instances)

	// 记录已经重试多少次了
	retried := 0
retry:

	// 如果没有开启RetrySameNode, 可以重新拿一个
	if retried != 0 {
		// 检查backoff策略
		if cli.ops.RetryPolocy.BackoffPolocy.Type == options.RetryBackoffPolicyTypeRangeTime {
			minTime := cli.ops.RetryPolocy.BackoffPolocy.TimeMin
			maxTime := cli.ops.RetryPolocy.BackoffPolocy.TimeMax
			time.Sleep(time.Duration(
				int64(minTime) +
					rand.Int63()%(int64(maxTime)-(int64(minTime))),
			))
		}

		// 换新节点
		if !cli.ops.RetryPolocy.RetrySameNode {
			cli.instanceMu.Lock()
			if len(cli.instances) == 0 {
				cli.instanceMu.Unlock()
				return nil, &clientCallError{
					IsNoInstance: true,
				}
			}
			ins = cli.ops.LoadBalance.GetNext(cli.instances)
			cli.instanceMu.Unlock()
		}
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
		now := time.Now()
		thisTimeDeadline := now.Add(cli.ops.Timeout)
		if retryEnabled {
			thisTimeDeadline = timeMinor(now.Add(cli.ops.Timeout), retryTotalDeadline)
		}
		conn, err := net.DialTimeout("tcp", ins.IPPort, thisTimeDeadline.Sub(now))
		if err != nil {
			networkError = err
			return
		}
		defer conn.Close()
		conn.SetDeadline(thisTimeDeadline)
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
		if retryEnabled && retried < retryPolicy.MaxRetryTimes && time.Now().Before(retryTotalDeadline) {
			retried++
			goto retry
		}
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
	IsNoInstance    bool

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
func (e *clientCallError) IsNoInstanceError() bool {
	return e.IsNoInstance
}
func (e *clientCallError) String() string {
	switch {
	case e.isBizError:
		return e.bizError.ErrXMessage()
	case e.isBusyError:
		return e.busyError.ErrXMessage()
	case e.isNetworkError:
		return e.networkError.Error()
	case e.isProtocolError:
		return "protocol error, check your proto defination!"
	case e.IsNoInstance:
		return "no instance found"
	}
	return "unexpected"
}
