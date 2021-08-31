package udprpc

type RpcType uint8
type RpcService uint16
type RpcId uint64
type RpcPayload []byte
type RpcError []byte

type RpcMessage struct {
	Type RpcType
	Service RpcService
	Id RpcId
	Payload RpcPayload
	Error RpcError
}

const RpcTypeRequest = RpcType(1)
const RpcTypeResponse = RpcType(2)
