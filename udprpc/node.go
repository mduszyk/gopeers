package udprpc

import (
	"errors"
	"log"
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
	services []RpcFunc
	addr *net.UDPAddr
	conn *net.UDPConn
	pendingRequests map[RpcId]*pendingRequest
	pendingMutex *sync.Mutex
	encoder Encoder
	decoder Decoder
}

func NewRpcNode(address string, services []RpcFunc) (*rpcNode, error) {
	n := &rpcNode{}
	n.services = services
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	n.addr = addr
	log.Printf("rpcNode addr: %v\n", n.addr)
	conn, err := net.ListenUDP("udp", n.addr)
	if err != nil {
		return nil, err
	}
	n.conn = conn
	n.pendingRequests = make(map[RpcId]*pendingRequest)
	n.pendingMutex = &sync.Mutex{}
	n.encoder = NewJsonEncoder()
	n.decoder = NewJsonDecoder()
	return n, nil
}

func (n *rpcNode) Run() {
	var message RpcMessage
	b := make([]byte, 0)
	for {
		_, addr, err := n.conn.ReadFromUDP(b)
		if err != nil {
			log.Printf("failed reading from udp, error: %s\n", err)
			continue
		}
		err = n.decoder.Decode(b, &message)
		if err != nil {
			log.Printf("failed decoding message, error: %s\n", err)
			continue
		}
		log.Printf("received message: %v, from: %v\n", message, addr)
		if message.Type == RpcTypeRequest {
			go n.handleRequest(message, addr)
		} else if message.Type == RpcTypeResponse {
			go n.handleResponse(message)
		} else {
			log.Printf("received unsupported message type: %v\n", message)
		}
	}
}

func (n *rpcNode) handleRequest(request RpcMessage, addr *net.UDPAddr) {
	fn := n.services[request.Service]
	result := fn(request.Payload)
	response := &RpcMessage{
		Type: RpcTypeResponse,
		Id: request.Id,
		Payload: result,
	}
	err := n.send(response, addr)
	if err != nil {
		log.Printf("failed handling request: %v, error: %s", request, err)
	}
}

func (n *rpcNode) handleResponse(response RpcMessage) {
	n.pendingMutex.Lock()
	if pending, ok := n.pendingRequests[response.Id]; ok {
		pending.response <- response
	} else {
		log.Printf("received unexpected response: %v\n", response)
	}
	n.pendingMutex.Unlock()
}

func (n *rpcNode) send(message *RpcMessage, addr *net.UDPAddr) error {
	buf, err := n.encoder.Encode(message)
	if err != nil {
		return err
	}
	_, err = n.conn.WriteToUDP(buf, addr)
	return err
}

func (n *rpcNode) addPending(id RpcId, pending *pendingRequest) {
	n.pendingMutex.Lock()
	n.pendingRequests[id] = pending
	n.pendingMutex.Unlock()
}

func (n *rpcNode) removePending(id RpcId) {
	n.pendingMutex.Lock()
	delete(n.pendingRequests, id)
	n.pendingMutex.Unlock()
}

func (n *rpcNode) Call(service RpcService, payload RpcPayload, addr *net.UDPAddr, timeout time.Duration) (*RpcMessage, error) {
	var id [8]byte // TODO
	request := &RpcMessage{
		Type: RpcTypeRequest,
		Service: service,
		Id: id,
		Payload: payload,
	}
	pending := &pendingRequest{request, make(chan RpcMessage, 1)}
	n.addPending(request.Id, pending)
	err := n.send(request, addr)
	if err != nil {
		return nil, err
	}
	select {
	case response := <-pending.response:
		n.removePending(request.Id)
		return &response, nil
	case <-time.After(timeout):
		n.removePending(request.Id)
		return nil, errors.New("timeout")
	}
}
