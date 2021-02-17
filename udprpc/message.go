package udprpc

const TypeService1Request = uint32(1)
const TypeService1Response = uint32(2)

type RpcMessage struct {
	Id [8]byte
	Type uint32
	Payload []byte
}
