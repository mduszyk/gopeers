package udprpc

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type RpcFunc func(payload RpcPayload) (RpcPayload, error)

type pendingRequest struct {
	request *RpcMessage
	response chan RpcMessage
}

type rpcNode struct {
	Addr            *net.UDPAddr
	services        []RpcFunc
	conn            *net.UDPConn
	pendingRequests map[RpcId]*pendingRequest
	pendingMutex    *sync.Mutex
	encoder         Encoder
	decoder         Decoder
	callTimeout     time.Duration
	readBufferSize  uint32
}

func NewRpcNode(address string, services []RpcFunc) (*rpcNode, error) {
	node := &rpcNode{
		callTimeout:     500 * time.Millisecond,
		readBufferSize:  1024,
		pendingRequests: make(map[RpcId]*pendingRequest),
		pendingMutex:    &sync.Mutex{},
		encoder:         NewJsonEncoder(),
		decoder:         NewJsonDecoder(),
		services:        services,
	}
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	addr, err = net.ResolveUDPAddr("udp", conn.LocalAddr().String())
	if err != nil {
		return nil, err
	}
	log.Printf("rpcNode addr: %v\n", addr)
	node.Addr = addr
	node.conn = conn
	return node, nil
}

func (node *rpcNode) Run() {
	var message RpcMessage
	buf := make([]byte, node.readBufferSize)
	for {
		n, addr, err := node.conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("failed reading from udp conn, error: %s\n", err)
			continue
		}
		err = node.decoder.Decode(buf[:n], &message)
		if err != nil {
			log.Printf("failed decoding message, error: %s\n", err)
			continue
		}
		switch message.Type {
		case RpcTypeRequest:
			go node.handleRequest(message, addr)
		case RpcTypeResponse:
			go node.handleResponse(message)
		default:
			log.Printf("received unsupported message type: %v\n", message)
		}
	}
}

func (node *rpcNode) handleRequest(request RpcMessage, addr *net.UDPAddr) {
	fn := node.services[request.Service]
	result, err := fn(request.Payload)
	response := &RpcMessage{
		Type: RpcTypeResponse,
		Id: request.Id,
		Payload: result,
	}
	if err != nil {
		response.Payload = nil
		response.Error = []byte(err.Error())
	}
	err = node.send(response, addr)
	if err != nil {
		log.Printf("failed sending response, request: %v, error: %s", request, err)
	}
}

func (node *rpcNode) handleResponse(response RpcMessage) {
	node.pendingMutex.Lock()
	if pending, ok := node.pendingRequests[response.Id]; ok {
		pending.response <- response
	} else {
		log.Printf("received unexpected response: %v\n", response)
	}
	node.pendingMutex.Unlock()
}

func (node *rpcNode) send(message *RpcMessage, addr *net.UDPAddr) error {
	buf, err := node.encoder.Encode(message)
	if err != nil {
		return err
	}
	n, err := node.conn.WriteToUDP(buf, addr)
	if err == nil && len(buf) != n {
		return errors.New("incomplete udp write")
	}
	return err
}

func (node *rpcNode) addPending(id RpcId, pending *pendingRequest) {
	node.pendingMutex.Lock()
	node.pendingRequests[id] = pending
	node.pendingMutex.Unlock()
}

func (node *rpcNode) removePending(id RpcId) {
	node.pendingMutex.Lock()
	delete(node.pendingRequests, id)
	node.pendingMutex.Unlock()
}

func (node *rpcNode) Call(addr *net.UDPAddr, service RpcService, payload RpcPayload) (RpcPayload, error) {
	request := &RpcMessage{
		Type: RpcTypeRequest,
		Service: service,
		Id: newId(),
		Payload: payload,
	}
	pending := &pendingRequest{request, make(chan RpcMessage, 1)}
	node.addPending(request.Id, pending)
	err := node.send(request, addr)
	if err != nil {
		return nil, err
	}
	select {
	case response := <-pending.response:
		node.removePending(request.Id)
		if response.Error != nil {
			return nil, errors.New(string(response.Error))
		} else {
			return response.Payload, nil
		}
	case <-time.After(node.callTimeout):
		node.removePending(request.Id)
		return nil, errors.New("call timeout")
	}
}
