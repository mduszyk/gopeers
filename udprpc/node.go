package udprpc

import (
	"log"
	"net"
	"sync"
)

type RpcFunc func(interface{}) interface{}

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
	return n, nil
}
