package udprpc

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type RpcPayload = []byte

type RpcError = []byte

type RpcId = uint64

type RpcService = uint32

type RpcFunc func(addr *net.UDPAddr, payload RpcPayload) (RpcPayload, error)

type pendingRequest struct {
	request *RpcMessage
	response chan RpcMessage
}

type RpcNode struct {
	Addr            *net.UDPAddr
	services        []RpcFunc
	conn            *net.UDPConn
	pendingRequests map[RpcId]*pendingRequest
	pendingMutex    *sync.Mutex
	callTimeout     time.Duration
	readBufferSize  uint32
	lastRpcId       uint64
}

func NewRpcNode(
	address string,
	services []RpcFunc,
	callTimeout time.Duration,
	readBufferSize uint32,
) (*RpcNode, error) {
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
	log.Printf("RpcNode addr: %v\n", addr)
	node := &RpcNode{
		callTimeout:     callTimeout,
		readBufferSize:  readBufferSize,
		pendingRequests: make(map[RpcId]*pendingRequest),
		pendingMutex:    &sync.Mutex{},
		services:        services,
		Addr:            addr,
		conn:            conn,
	}
	return node, nil
}

func (node *RpcNode) Run() {
	var message RpcMessage
	buf := make([]byte, node.readBufferSize)
	for {
		n, addr, err := node.conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("failed reading from udp conn, error: %s\n", err)
			continue
		}
		err = proto.Unmarshal(buf[:n], &message)
		if err != nil {
			log.Printf("failed decoding message: %s, error: %s\n", string(buf[:n]), err)
			continue
		}
		switch message.Type {
		case RpcMessage_REQUEST:
			go node.handleRequest(message, addr)
		case RpcMessage_RESPONSE:
			go node.handleResponse(message)
		default:
			log.Printf("received unsupported message type: %v\n", message)
		}
	}
}

func (node *RpcNode) handleRequest(request RpcMessage, addr *net.UDPAddr) {
	fn := node.services[request.Service]
	result, err := fn(addr, request.Payload)
	response := &RpcMessage{
		Type: RpcMessage_RESPONSE,
		RpcId: request.RpcId,
		Payload: result,
	}
	if err != nil {
		response.Payload = nil
		response.Error = RpcError(err.Error())
	}
	err = node.send(response, addr)
	if err != nil {
		log.Printf("failed sending response, request: %v, error: %s", request, err)
	}
}

func (node *RpcNode) handleResponse(response RpcMessage) {
	node.pendingMutex.Lock()
	if pending, ok := node.pendingRequests[response.RpcId]; ok {
		pending.response <- response
	} else {
		log.Printf("received unexpected response: %v\n", response)
	}
	node.pendingMutex.Unlock()
}

func (node *RpcNode) send(message *RpcMessage, addr *net.UDPAddr) error {
	buf, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	n, err := node.conn.WriteToUDP(buf, addr)
	if err == nil && len(buf) != n {
		return errors.New("incomplete udp write")
	}
	return err
}

func (node *RpcNode) addPending(id RpcId, pending *pendingRequest) {
	node.pendingMutex.Lock()
	node.pendingRequests[id] = pending
	node.pendingMutex.Unlock()
}

func (node *RpcNode) removePending(id RpcId) {
	node.pendingMutex.Lock()
	delete(node.pendingRequests, id)
	node.pendingMutex.Unlock()
}

func (node *RpcNode) nextRpcId() RpcId {
	return atomic.AddUint64(&node.lastRpcId, 1)
}

func (node *RpcNode) Call(addr *net.UDPAddr, service RpcService, payload RpcPayload) (RpcPayload, error) {
	request := &RpcMessage{
		Type:    RpcMessage_REQUEST,
		Service: service,
		RpcId:      node.nextRpcId(),
		Payload: payload,
	}
	pending := &pendingRequest{request, make(chan RpcMessage, 1)}
	node.addPending(request.RpcId, pending)
	err := node.send(request, addr)
	if err != nil {
		return nil, err
	}
	select {
	case response := <-pending.response:
		node.removePending(request.RpcId)
		if response.Error != nil {
			return nil, errors.New(string(response.Error))
		} else {
			return response.Payload, nil
		}
	case <-time.After(node.callTimeout):
		node.removePending(request.RpcId)
		return nil, errors.New("call timeout")
	}
}
