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

func NewUdpProtocolServer(address string, p2pNode *p2pNode) (*udpProtocolServer, error) {
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
	rpcNode, err := udprpc.NewRpcNode(address, services)
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

type pingRequest struct {
	PeerAddr *net.UDPAddr
	PeerId   Id
	RandomId Id
}

type pingResponse struct {
	RandomId Id
}

type findNodeRequest struct {
	PeerAddr *net.UDPAddr
	PeerId   Id
	NodeId Id
}

type udpNode struct {
	Addr *net.UDPAddr
	NodeId Id
}

type findNodeResponse struct {
	Nodes 	 []udpNode
}

func (s *udpProtocolServer) PingRpc(requestPayload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	var pingRequest pingRequest
	err := s.decoder.Decode(requestPayload, &pingRequest)
	if err != nil {
		return nil, err
	}
	peer := &Peer{Id: pingRequest.PeerId, LastSeen: time.Now()}
	s.Connect(pingRequest.PeerAddr, peer)
	pingId, err := s.p2pNode.Ping(peer, pingRequest.RandomId)
	if err != nil {
		return nil, err
	}
	pingResult := pingResponse{RandomId: pingId}
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

func (p *udpProtocol) Ping(_ *Peer, randomId Id) (Id, error) {
	pingRequest := pingRequest{
		PeerAddr: p.server.rpcNode.Addr,
		PeerId: p.server.p2pNode.peer.Id,
		RandomId: randomId,
	}
	requestPayload, err := p.server.encoder.Encode(&pingRequest)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.server.rpcNode.Call(p.addr, p.server.pingService, requestPayload)
	if err != nil {
		return nil, err
	}
	var pingResponse pingResponse
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
