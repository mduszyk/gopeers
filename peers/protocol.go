package peers

import (
	"github.com/golang/protobuf/proto"
	"github.com/mduszyk/gopeers/udprpc"
	"net"
	"time"
)

type FindResult struct {
	peers []*Peer
	value []byte
}

type Protocol interface {
	Ping(sender *Peer, randomId Id) (Id, error)
	FindNode(sender *Peer, id Id) ([]*Peer, error)
	FindValue(sender *Peer, id Id) (*FindResult, error)
	Store(sender *Peer, id Id, value []byte) error
}

type udpProtocolServer struct {
	rpcNode *udprpc.RpcNode
	p2pNode *P2pNode
	pingService udprpc.RpcService
	findNodeService udprpc.RpcService
}

func NewUdpProtocolServer(
	address string,
	p2pNode *P2pNode,
	rpcCallTimeout time.Duration,
	readBufferSize uint32) (*udpProtocolServer, error) {

	protoServer :=  &udpProtocolServer{
		p2pNode: p2pNode,
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

func (s *udpProtocolServer) PingRpc(addr *net.UDPAddr, payload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	var request PingRequest
	err := proto.Unmarshal(payload, &request)
	if err != nil {
		return nil, err
	}
	peer := &Peer{Id: BytesId(request.PeerId), LastSeen: time.Now()}
	s.Connect(addr, peer)
	pingId, err := s.p2pNode.Ping(peer, BytesId(request.RandomId))
	if err != nil {
		return nil, err
	}
	response := PingResponse{RandomId: pingId.Bytes()}
	return proto.Marshal(&response)
}

func (s *udpProtocolServer) FindNodeRpc(addr *net.UDPAddr, payload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	var request FindNodeRequest
	err := proto.Unmarshal(payload, &request)
	if err != nil {
		return nil, err
	}
	peer := &Peer{Id: BytesId(request.PeerId), LastSeen: time.Now()}
	s.Connect(addr, peer)
	peers, err := s.p2pNode.FindNode(peer, BytesId(request.NodeId))
	if err != nil {
		return nil, err
	}
	nodes := make([]*UdpNode, len(peers))
	for i := 0; i< len(peers); i++ {
		protocol := (peers[i].Proto).(*udpProtocol)
		protoAddr := &UDPAddr{
			IP:   protocol.addr.IP,
			Port: int32(protocol.addr.Port),
			Zone: protocol.addr.Zone,
		}
		nodes[i] = &UdpNode{Addr: protoAddr, NodeId: peers[i].Id.Bytes()}
	}
	response := FindNodeResponse{Nodes: nodes}
	return proto.Marshal(&response)
}

func (s *udpProtocolServer) FindValueRpc(addr *net.UDPAddr, payload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	// TODO
	return nil, nil
}

func (s *udpProtocolServer) StoreRpc(addr *net.UDPAddr, payload udprpc.RpcPayload) (udprpc.RpcPayload, error) {
	// TODO
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
	request := PingRequest{
		PeerId: p.server.p2pNode.Peer.Id.Bytes(),
		RandomId: randomId.Bytes(),
	}
	requestPayload, err := proto.Marshal(&request)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.server.rpcNode.Call(p.addr, p.server.pingService, requestPayload)
	if err != nil {
		return nil, err
	}
	var response PingResponse
	err = proto.Unmarshal(responsePayload, &response)
	if err != nil {
		return nil, err
	}
	return BytesId(response.RandomId), nil
}

func (p *udpProtocol) FindNode(_ *Peer, id Id) ([]*Peer, error) {
	request := FindNodeRequest{
		PeerId: p.server.p2pNode.Peer.Id.Bytes(),
		NodeId: id.Bytes(),
	}
	requestPayload, err := proto.Marshal(&request)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.server.rpcNode.Call(p.addr, p.server.findNodeService, requestPayload)
	if err != nil {
		return nil, err
	}
	var response FindNodeResponse
	err = proto.Unmarshal(responsePayload, &response)
	if err != nil {
		return nil, err
	}
	peers := make([]*Peer, len(response.Nodes))
	for i := 0; i < len(response.Nodes); i++ {
		n := response.Nodes[i]
		peer := &Peer{Id: BytesId(n.NodeId), LastSeen: time.Now()}
		addr := &net.UDPAddr{
			IP: n.Addr.IP,
			Port: int(n.Addr.Port),
			Zone: n.Addr.Zone,
		}
		p.server.Connect(addr, peer)
		peers[i] = peer
	}
	return peers, nil
}


func (p *udpProtocol) FindValue(sender *Peer, id Id) (*FindResult, error) {
	// TODO
	return nil, nil
}

func (p *udpProtocol) Store(sender *Peer, id Id, value []byte) error {
	// TODO
	return nil
}

func NewUdpProtoNode(k, b int, address string,
rpcCallTimeout time.Duration, rpcReadBufferSize uint32) (*udpProtocolServer, error) {
	nodeId, err := CryptoRandId()
	if err != nil {
		return nil, err
	}
	node := NewP2pNode(k, b, nodeId)
	return NewUdpProtocolServer(address, node, rpcCallTimeout, rpcReadBufferSize)
}
