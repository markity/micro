package protocol

/*
消息协议定义:

请求:
4字节调用名字的长度 + 4字节proto buf的长度 + 调用名字bytes + proto buf bytes

响应:
1字节错误代码定义 + 4字节errcode长度 + 4字节proto buf长度 + errcode gob 编码的 bytes + proto buf bytes
任何一个长度为0, 则不需要对应bytes, 此时客户端那边这个字段是nil

如果发生协议错误, 有两种, 一是没有这个HandleName, 二是proto unmarshal错误, 需要区分这两种类型
*/

// 协议:
// 请求: 4字节代表调用名字, 4字节代表proto body的长度
// 响应: 1字节错误代码, 4字节proto body长度(仅当错误代码为0时)
// 错误代码定义: 0(success) 1(handleName不存在) 2(解析请求proto错误) 3(有error)

type ProtocolErrorType byte

// 错误代码为1或2时 4字节errcode长度 + 4字节proto buf长度 均为 0
var ProtocolErrorTypeSuccess ProtocolErrorType = 0
var ProtocolErrorTypeHandleNameInvalid ProtocolErrorType = 1
var ProtocolErrorTypeParseProtoFailed ProtocolErrorType = 2
var ProtocolErrorUnexpected ProtocolErrorType = 3
