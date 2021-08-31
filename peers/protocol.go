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
	p2pNode *P2pNode
	encoder udprpc.Encoder
	decoder udprpc.Decoder
	pingService udprpc.RpcService
	findNodeService udprpc.RpcService
}

func NewUdpProtocolServer(
	address string,
	p2pNode *P2pNode,
	rpcCallTimeout time.Duration,
	readBufferSize uint32) (*udpProtocolServer, error) {

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
	rpcNode, err := udprpc.NewRpcNode(address, services, rpcCallTimeout, readBufferSize)
	if err != nil {
		return nil, err
	}
	protoServer.rpcNode = rpcNode
	go rpcNode.Run()
	return protoServer, nil
}

func (s *udpProtocolServer) Connect(peerAddr *net.UDPAddr, peer *Peer) {
	peer.Proto = NewUdpProtocol(peerAddr, s)
}

type pingRequest struct {
	PeerId   Id
	RandomId Id
}

type pingResponse struct {
	RandomId Id
}

type findNodeRequest struct {
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

func (s *udpProtocolServer) PingRpc(addr *net.UDPAddr, payload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	var request pingRequest
	err := s.decoder.Decode(payload, &request)
	if err != nil {
		return nil, err
	}
	peer := &Peer{Id: request.PeerId, LastSeen: time.Now()}
	s.Connect(addr, peer)
	pingId, err := s.p2pNode.Ping(peer, request.RandomId)
	if err != nil {
		return nil, err
	}
	response := pingResponse{RandomId: pingId}
	return s.encoder.Encode(&response)
}

func (s *udpProtocolServer) FindNodeRpc(addr *net.UDPAddr, payload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	var request findNodeRequest
	err := s.decoder.Decode(payload, &request)
	if err != nil {
		return nil, err
	}
	peer := &Peer{Id: request.PeerId, LastSeen: time.Now()}
	s.Connect(addr, peer)
	peers, err := s.p2pNode.FindNode(peer, request.NodeId)
	if err != nil {
		return nil, err
	}
	nodes := make([]udpNode, len(peers))
	for i := 0; i< len(peers); i++ {
		proto := (peers[i].Proto).(*udpProtocol)
		nodes[i] = udpNode{Addr: proto.addr, NodeId: peers[i].Id}
	}
	response := findNodeResponse{Nodes: nodes}
	return s.encoder.Encode(&response)
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
	request := pingRequest{
		PeerId: p.server.p2pNode.Peer.Id,
		RandomId: randomId,
	}
	requestPayload, err := p.server.encoder.Encode(&request)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.server.rpcNode.Call(p.addr, p.server.pingService, requestPayload)
	if err != nil {
		return nil, err
	}
	var response pingResponse
	err = p.server.decoder.Decode(responsePayload, &response)
	if err != nil {
		return nil, err
	}
	return response.RandomId, nil
}

func (p *udpProtocol) FindNode(_ *Peer, id Id) ([]*Peer, error) {
	request := findNodeRequest{
		PeerId: p.server.p2pNode.Peer.Id,
		NodeId: id,
	}
	requestPayload, err := p.server.encoder.Encode(&request)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.server.rpcNode.Call(p.addr, p.server.findNodeService, requestPayload)
	if err != nil {
		return nil, err
	}
	var response findNodeResponse
	err = p.server.decoder.Decode(responsePayload, &response)
	if err != nil {
		return nil, err
	}
	peers := make([]*Peer, len(response.Nodes))
	for i := 0; i < len(response.Nodes); i++ {
		n := response.Nodes[i]
		peer := &Peer{Id: n.NodeId, LastSeen: time.Now()}
		p.server.Connect(n.Addr, peer)
		peers[i] = peer
	}
	return peers, nil
}

func NewUdpProtoNode(k, b int, address string,
rpcCallTimeout time.Duration, rpcReadBufferSize uint32) (*udpProtocolServer, error) {
	nodeId, err := CryptoRandId()
	if err != nil {
		return nil, err
	}
	node := NewP2pNode(k, b, nodeId)
	protoServer, err := NewUdpProtocolServer(address, node, rpcCallTimeout, rpcReadBufferSize)
	return protoServer, nil
}
