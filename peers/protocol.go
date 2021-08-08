package peers

import (
	"github.com/mduszyk/gopeers/udprpc"
	"net"
	"time"
)

type Protocol interface {
	Ping(sender *Peer, randomId Id) (Id, error)
	FindNode(sender *Peer) error
}

type udpProtocolServer struct {
	rpcNode *udprpc.RpcNode
	p2pNode *p2pNode
	encoder udprpc.Encoder
	decoder udprpc.Decoder
	pingService udprpc.RpcService
}

func NewUdpProtocolServer(addres string, p2pNode *p2pNode) (*udpProtocolServer, error) {
	encoder := udprpc.NewJsonEncoder()
	decoder := udprpc.NewJsonDecoder()
	protoServer :=  &udpProtocolServer{
		p2pNode: p2pNode,
		encoder: encoder,
		decoder: decoder,
		pingService: udprpc.RpcService(0),
	}
	services := []udprpc.RpcFunc{protoServer.PingRpc}
	rpcNode, err := udprpc.NewRpcNode(addres, services)
	if err != nil {
		return nil, err
	}
	protoServer.rpcNode = rpcNode
	go rpcNode.Run()
	return protoServer, nil
}

type pingPayload struct {
	Addr     *net.UDPAddr
	Id       Id
	RandomId Id
}

func (s *udpProtocolServer) Connect(addr *net.UDPAddr, peer *Peer) {
	peer.Proto = NewUdpProtocol(addr, s)
}

func (s *udpProtocolServer) PingRpc(payload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	var req pingPayload
	err := s.decoder.Decode(payload, &req)
	if err != nil {
		return nil, err
	}
	sender := &Peer{Id: req.Id, LastSeen: time.Now()}
	s.Connect(req.Addr, sender)
	id, err := s.p2pNode.Ping(sender, req.RandomId)
	if err != nil {
		return nil, err
	}
	pingResult := pingPayload{s.rpcNode.Addr, s.p2pNode.peer.Id, id}
	response, err := s.encoder.Encode(&pingResult)
	return response, err
}

type udpProtocol struct {
	addr *net.UDPAddr
	server *udpProtocolServer
}

func NewUdpProtocol(addr *net.UDPAddr, server *udpProtocolServer) *udpProtocol {
	return &udpProtocol{
		addr: addr,
		server: server,
	}
}

func (p *udpProtocol) Ping(sender *Peer, randomId Id) (Id, error) {
	pingReq := pingPayload{
		p.server.rpcNode.Addr,
		p.server.p2pNode.peer.Id,
		randomId,
	}
	req, err := p.server.encoder.Encode(&pingReq)
	if err != nil {
		return nil, err
	}
	response, err := p.server.rpcNode.Call(p.addr, p.server.pingService, req)
	if err != nil {
		return nil, err
	}
	var payload pingPayload
	err = p.server.decoder.Decode(response, &payload)
	if err != nil {
		return nil, err
	}
	return payload.RandomId, nil
}

func (p *udpProtocol) FindNode(sender *Peer) error {
	return nil
}


func NewUdpProtoNode(k, b int, address string) (*Peer, *udpProtocolServer, error) {
	nodeId, err := RandomId()
	if err != nil {
		return nil, nil, err
	}
	peer := &Peer{Id: nodeId, LastSeen: time.Now()}
	node := NewP2pNode(k, b, peer)
	protoServer, err := NewUdpProtocolServer(address, node)
	return peer, protoServer, nil
}

func NewMethodCallProtoNode(k, b int) (*Peer, *p2pNode, error) {
	nodeId, err := RandomId()
	if err != nil {
		return nil, nil, err
	}
	peer := &Peer{Id: nodeId, LastSeen: time.Now()}
	node := NewP2pNode(k, b, peer)
	peer.Proto = node
	return peer, node, nil
}