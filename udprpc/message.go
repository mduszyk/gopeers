package udprpc

import "math/rand"

type RpcType uint32
type RpcService uint32
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

func randRpcId() RpcId {
	return RpcId(rand.Uint64())
}
