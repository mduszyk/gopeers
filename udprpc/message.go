package udprpc

const TypeEchoRequest = RpcType(1)
const TypeEchoResponse = RpcType(2)

type RpcType uint32
type RpcId [8]byte
type RpcPayload []byte

type RpcMessage struct {
	Type RpcType
	Id RpcId
	Payload RpcPayload
}
