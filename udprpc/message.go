package udprpc

const TypeEchoRequest = uint32(1)
const TypeEchoResponse = uint32(2)

type RpcMessage struct {
	Type uint32
	Id [8]byte
	Payload []byte
}
