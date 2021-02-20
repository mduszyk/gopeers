package udprpc

type RpcType uint32
type RpcService uint32
type RpcId uint64
type RpcPayload []byte

type RpcMessage struct {
	Type RpcType
	Service RpcService
	Id RpcId
	Payload RpcPayload
}

const RpcTypeRequest = RpcType(1)
const RpcTypeResponse = RpcType(2)

const RpcServiceEcho = RpcService(1)

