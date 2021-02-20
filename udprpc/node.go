package udprpc

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type RpcFunc func(payload RpcPayload) RpcPayload

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
	node.conn = conn
	addr, err = net.ResolveUDPAddr("udp", conn.LocalAddr().String())
	if err != nil {
		return nil, err
	}
	node.Addr = addr
	log.Printf("rpcNode addr: %v\n", node.Addr)
	return node, nil
}

func (node *rpcNode) Run() {
	var message RpcMessage
	buf := make([]byte, node.readBufferSize)
	for {
		n, addr, err := node.conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("failed reading from udp, error: %s\n", err)
			continue
		}
		err = node.decoder.Decode(buf[:n], &message)
		if err != nil {
			log.Printf("failed decoding message, error: %s\n", err)
			continue
		}
		log.Printf("received message: %v, from: %v\n", message, addr)
		if message.Type == RpcTypeRequest {
			go node.handleRequest(message, addr)
		} else if message.Type == RpcTypeResponse {
			go node.handleResponse(message)
		} else {
			log.Printf("received unsupported message type: %v\n", message)
		}
	}
}

func (node *rpcNode) handleRequest(request RpcMessage, addr *net.UDPAddr) {
	fn := node.services[request.Service]
	result := fn(request.Payload)
	response := &RpcMessage{
		Type: RpcTypeResponse,
		Id: request.Id,
		Payload: result,
	}
	err := node.send(response, addr)
	if err != nil {
		log.Printf("failed handling request: %v, error: %s", request, err)
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
	_, err = node.conn.WriteToUDP(buf, addr)
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
	id := RpcId(rand.Uint64()) // TODO
	request := &RpcMessage{
		Type: RpcTypeRequest,
		Service: service,
		Id: id,
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
		return response.Payload, nil
	case <-time.After(node.callTimeout):
		node.removePending(request.Id)
		return nil, errors.New("timeout")
	}
}
