package peers

import (
	"github.com/mduszyk/gopeers/udprpc"
	"net"
	"time"
)

type Protocol interface {
	Ping(sender *Peer, randomId Id) (Id, error)
	FindNode(sender *Peer, id Id) ([]*Peer, error)
}

type udpProtocolServer struct {
	rpcNode *udprpc.RpcNode
	p2pNode *p2pNode
	encoder udprpc.Encoder
	decoder udprpc.Decoder
	pingService udprpc.RpcService
	findNodeService udprpc.RpcService
}

func NewUdpProtocolServer(addres string, p2pNode *p2pNode) (*udpProtocolServer, error) {
	encoder := udprpc.NewJsonEncoder()
	decoder := udprpc.NewJsonDecoder()
	protoServer :=  &udpProtocolServer{
		p2pNode: p2pNode,
		encoder: encoder,
		decoder: decoder,
		pingService: udprpc.RpcService(0),
		findNodeService: udprpc.RpcService(1),
	}
	services := []udprpc.RpcFunc{
		protoServer.PingRpc,
		protoServer.FindNodeRpc,
	}
	rpcNode, err := udprpc.NewRpcNode(addres, services)
	if err != nil {
		return nil, err
	}
	protoServer.rpcNode = rpcNode
	go rpcNode.Run()
	return protoServer, nil
}

func (s *udpProtocolServer) Connect(addr *net.UDPAddr, peer *Peer) {
	peer.Proto = NewUdpProtocol(addr, s)
}

type pingPayload struct {
	Addr     *net.UDPAddr
	Id       Id
	RandomId Id
}

func (s *udpProtocolServer) PingRpc(requestPayload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	var pingRequest pingPayload
	err := s.decoder.Decode(requestPayload, &pingRequest)
	if err != nil {
		return nil, err
	}
	sender := &Peer{Id: pingRequest.Id, LastSeen: time.Now()}
	s.Connect(pingRequest.Addr, sender)
	pingId, err := s.p2pNode.Ping(sender, pingRequest.RandomId)
	if err != nil {
		return nil, err
	}
	pingResult := pingPayload{s.rpcNode.Addr, s.p2pNode.peer.Id, pingId}
	responsePayload, err := s.encoder.Encode(&pingResult)
	return responsePayload, err
}

func (s *udpProtocolServer) FindNodeRpc(requestPayload udprpc.RpcPayload) (udprpc.RpcPayload, error) {

	//peers, err := s.p2pNode.FindNode(sender, findNodeRequest.FindId)

	return nil, nil
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
	pingRequest := pingPayload{
		p.server.rpcNode.Addr,
		p.server.p2pNode.peer.Id,
		randomId,
	}
	requestPayload, err := p.server.encoder.Encode(&pingRequest)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.server.rpcNode.Call(p.addr, p.server.pingService, requestPayload)
	if err != nil {
		return nil, err
	}
	var pingResponse pingPayload
	err = p.server.decoder.Decode(responsePayload, &pingResponse)
	if err != nil {
		return nil, err
	}
	return pingResponse.RandomId, nil
}

func (p *udpProtocol) FindNode(sender *Peer, id Id) ([]*Peer, error) {

	//responsePayload, err := p.server.rpcNode.Call(p.addr, p.server.findNodeService, requestPayload)

	return nil, nil
}


func NewUdpProtoNode(k, b int, address string) (*udpProtocolServer, error) {
	nodeId, err := RandomId()
	if err != nil {
		return nil, err
	}
	node := NewP2pNode(k, b, nodeId)
	protoServer, err := NewUdpProtocolServer(address, node)
	return protoServer, nil
}
