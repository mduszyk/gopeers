package rpc

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Payload = []byte

type Error = []byte

type CallId = uint64

type ServiceId = uint32

type Service func(addr *net.UDPAddr, payload Payload) (Payload, error)

type pendingCall struct {
	request *Message
	response chan Message
}

type UdpNode struct {
	Addr            *net.UDPAddr
	services        []Service
	conn            *net.UDPConn
	pendingRequests map[CallId]*pendingCall
	pendingMutex    *sync.RWMutex
	callTimeout     time.Duration
	readBufferSize  uint32
	lastCallId      uint64
}

func NewUdpNode(
	address string,
	services []Service,
	callTimeout time.Duration,
	readBufferSize uint32,
) (*UdpNode, error) {
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
	log.Printf("UdpNode addr: %v\n", addr)
	node := &UdpNode{
		callTimeout:     callTimeout,
		readBufferSize:  readBufferSize,
		pendingRequests: make(map[CallId]*pendingCall),
		pendingMutex:    &sync.RWMutex{},
		services:        services,
		Addr:            addr,
		conn:            conn,
	}
	return node, nil
}

func (node *UdpNode) Run() {
	var message Message
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
		case Message_REQUEST:
			go node.handleRequest(message, addr)
		case Message_RESPONSE:
			go node.handleResponse(message)
		default:
			log.Printf("received unsupported message type: %v\n", message)
		}
	}
}

func (node *UdpNode) handleRequest(request Message, addr *net.UDPAddr) {
	service := node.services[request.ServiceId]
	result, err := service(addr, request.Payload)
	response := &Message{
		Type: Message_RESPONSE,
		CallId: request.CallId,
		Payload: result,
	}
	if err != nil {
		response.Payload = nil
		response.Error = Error(err.Error())
	}
	err = node.send(response, addr)
	if err != nil {
		log.Printf("failed sending response, request: %v, error: %s", request, err)
	}
}

func (node *UdpNode) handleResponse(response Message) {
	node.pendingMutex.RLock()
	pending, ok := node.pendingRequests[response.CallId]
	node.pendingMutex.RUnlock()
	if ok {
		pending.response <- response
	} else {
		log.Printf("received unexpected response: %v\n", response)
	}
}

func (node *UdpNode) send(message *Message, addr *net.UDPAddr) error {
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

func (node *UdpNode) addPending(id CallId, pending *pendingCall) {
	node.pendingMutex.Lock()
	node.pendingRequests[id] = pending
	node.pendingMutex.Unlock()
}

func (node *UdpNode) removePending(id CallId) {
	node.pendingMutex.Lock()
	delete(node.pendingRequests, id)
	node.pendingMutex.Unlock()
}

func (node *UdpNode) nextCallId() CallId {
	return atomic.AddUint64(&node.lastCallId, 1)
}

func (node *UdpNode) Call(addr *net.UDPAddr, serviceId ServiceId, payload Payload) (Payload, error) {
	request := &Message{
		Type:      Message_REQUEST,
		ServiceId: serviceId,
		CallId:    node.nextCallId(),
		Payload:   payload,
	}
	pending := &pendingCall{request, make(chan Message, 1)}
	node.addPending(request.CallId, pending)
	err := node.send(request, addr)
	if err != nil {
		return nil, err
	}
	select {
	case response := <-pending.response:
		node.removePending(request.CallId)
		if response.Error != nil {
			return nil, errors.New(string(response.Error))
		} else {
			return response.Payload, nil
		}
	case <-time.After(node.callTimeout):
		node.removePending(request.CallId)
		return nil, errors.New("call timeout")
	}
}
