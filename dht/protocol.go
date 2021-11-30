package dht

import (
	"github.com/golang/protobuf/proto"
	"github.com/mduszyk/gopeers/rpc"
	"github.com/mduszyk/gopeers/store"
	"net"
	"time"
)

type FindResult struct {
	peers []*Peer
	value []byte
}

type Protocol interface {
	Ping(sender *Peer, randomId Id) (Id, error)
	FindNode(sender *Peer, id Id) (*FindResult, error)
	FindValue(sender *Peer, key Id) (*FindResult, error)
	Store(sender *Peer, key Id, value []byte) error
}

type udpProtocolNode struct {
	rpcNode            *rpc.UdpNode
	dhtNode            *KadNode
	pingServiceId      rpc.ServiceId
	findNodeServiceId  rpc.ServiceId
	findValueServiceId rpc.ServiceId
	storeServiceId     rpc.ServiceId
}

func NewUdpProtocolNode(rpcNode *rpc.UdpNode, dhtNode *KadNode) *udpProtocolNode {
	protocolNode := &udpProtocolNode{
		rpcNode:            rpcNode,
		dhtNode:            dhtNode,
		pingServiceId:      rpc.ServiceId(0),
		findNodeServiceId:  rpc.ServiceId(1),
		findValueServiceId: rpc.ServiceId(2),
		storeServiceId:     rpc.ServiceId(3),
	}
	// register rpc services
	rpcNode.Services = []rpc.Service{
		protocolNode.PingRpc,
		protocolNode.FindNodeRpc,
		protocolNode.FindValueRpc,
		protocolNode.StoreRpc,
	}
	return protocolNode
}

func StartUdpProtocolNode(
	k, b, alpha int,
	nodeId Id,
	storage store.Storage,
	address string,
	rpcCallTimeout time.Duration,
	readBufferSize uint32,
	) (*udpProtocolNode, error) {

	dhtNode := NewKadNode(k, b, alpha, nodeId, storage)
	rpcNode, err := rpc.NewUdpNode(address, nil, rpcCallTimeout, readBufferSize)
	if err != nil {
		return nil, err
	}
	protocolNode := NewUdpProtocolNode(rpcNode, dhtNode)

	go rpcNode.Run()

	return protocolNode, nil
}

func (n *udpProtocolNode) Connect(peerAddr *net.UDPAddr, peer *Peer) {
	peer.Proto = NewUdpProtocol(peerAddr, n)
}

func (n *udpProtocolNode) PingRpc(addr *net.UDPAddr, payload rpc.Payload) (rpc.Payload, error) {
	var request PingRequest
	err := proto.Unmarshal(payload, &request)
	if err != nil {
		return nil, err
	}
	peer := NewPeer(BytesId(request.PeerId))
	n.Connect(addr, peer)
	pingId, err := n.dhtNode.Ping(peer, BytesId(request.RandomId))
	if err != nil {
		return nil, err
	}
	response := PingResponse{RandomId: pingId.Bytes()}
	return proto.Marshal(&response)
}

func (n *udpProtocolNode) FindNodeRpc(addr *net.UDPAddr, payload rpc.Payload) (rpc.Payload, error) {
	var request FindRequest
	err := proto.Unmarshal(payload, &request)
	if err != nil {
		return nil, err
	}
	peer := NewPeer(BytesId(request.PeerId))
	n.Connect(addr, peer)
	findResult, err := n.dhtNode.FindNode(peer, BytesId(request.PeerId))
	if err != nil {
		return nil, err
	}
	nodes := make([]*UdpNode, len(findResult.peers))
	for i := 0; i< len(findResult.peers); i++ {
		protocol := (findResult.peers[i].Proto).(*udpProtocol)
		protoAddr := &UDPAddr{
			IP:   protocol.addr.IP,
			Port: int32(protocol.addr.Port),
			Zone: protocol.addr.Zone,
		}
		nodes[i] = &UdpNode{Addr: protoAddr, NodeId: findResult.peers[i].Id.Bytes()}
	}
	response := FindNodeResponse{Nodes: nodes}
	return proto.Marshal(&response)
}

func (n *udpProtocolNode) FindValueRpc(addr *net.UDPAddr, payload rpc.Payload) (rpc.Payload, error) {
	var request FindRequest
	err := proto.Unmarshal(payload, &request)
	if err != nil {
		return nil, err
	}
	peer := NewPeer(BytesId(request.PeerId))
	n.Connect(addr, peer)
	findResult, err := n.dhtNode.FindValue(peer, BytesId(request.Id))
	if err != nil {
		return nil, err
	}
	var nodes []*UdpNode
	if findResult.peers != nil {
		nodes = make([]*UdpNode, len(findResult.peers))
		for i := 0; i < len(findResult.peers); i++ {
			protocol := (findResult.peers[i].Proto).(*udpProtocol)
			protoAddr := &UDPAddr{
				IP:   protocol.addr.IP,
				Port: int32(protocol.addr.Port),
				Zone: protocol.addr.Zone,
			}
			nodes[i] = &UdpNode{Addr: protoAddr, NodeId: findResult.peers[i].Id.Bytes()}
		}
	}
	response := FindValueResponse{Nodes: nodes, Value: findResult.value}
	return proto.Marshal(&response)
}

func (n *udpProtocolNode) StoreRpc(addr *net.UDPAddr, payload rpc.Payload) (rpc.Payload, error) {
	var request StoreRequest
	err := proto.Unmarshal(payload, &request)
	if err != nil {
		return nil, err
	}
	peer := NewPeer(BytesId(request.PeerId))
	n.Connect(addr, peer)
	err = n.dhtNode.Store(peer, BytesId(request.Key), request.Value)
	if err != nil {
		return nil, err
	}
	return nil, err
}

type udpProtocol struct {
	addr         *net.UDPAddr
	protocolNode *udpProtocolNode
}

func NewUdpProtocol(addr *net.UDPAddr, server *udpProtocolNode) *udpProtocol {
	return &udpProtocol{
		addr:         addr,
		protocolNode: server,
	}
}

func (p *udpProtocol) Ping(_ *Peer, randomId Id) (Id, error) {
	request := PingRequest{
		PeerId: p.protocolNode.dhtNode.Peer.Id.Bytes(),
		RandomId: randomId.Bytes(),
	}
	requestPayload, err := proto.Marshal(&request)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.protocolNode.rpcNode.Call(p.addr, p.protocolNode.pingServiceId, requestPayload)
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

func (p *udpProtocol) FindNode(_ *Peer, id Id) (*FindResult, error) {
	request := FindRequest{
		PeerId: p.protocolNode.dhtNode.Peer.Id.Bytes(),
		Id: id.Bytes(),
	}
	requestPayload, err := proto.Marshal(&request)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.protocolNode.rpcNode.Call(p.addr, p.protocolNode.findNodeServiceId, requestPayload)
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
		p.protocolNode.Connect(addr, peer)
		peers[i] = peer
	}
	result := &FindResult{peers: peers, value: nil}
	return result, nil
}


func (p *udpProtocol) FindValue(sender *Peer, key Id) (*FindResult, error) {
	request := FindRequest{
		PeerId: p.protocolNode.dhtNode.Peer.Id.Bytes(),
		Id: key.Bytes(),
	}
	requestPayload, err := proto.Marshal(&request)
	if err != nil {
		return nil, err
	}
	responsePayload, err := p.protocolNode.rpcNode.Call(
		p.addr, p.protocolNode.findValueServiceId, requestPayload)
	if err != nil {
		return nil, err
	}
	var response FindValueResponse
	err = proto.Unmarshal(responsePayload, &response)
	if err != nil {
		return nil, err
	}
	var peers []*Peer
	if response.Nodes != nil {
		peers = make([]*Peer, len(response.Nodes))
		for i := 0; i < len(response.Nodes); i++ {
			n := response.Nodes[i]
			peer := &Peer{Id: BytesId(n.NodeId), LastSeen: time.Now()}
			addr := &net.UDPAddr{
				IP:   n.Addr.IP,
				Port: int(n.Addr.Port),
				Zone: n.Addr.Zone,
			}
			p.protocolNode.Connect(addr, peer)
			peers[i] = peer
		}
	}
	result := &FindResult{peers: peers, value: response.Value}
	return result, nil
}

func (p *udpProtocol) Store(sender *Peer, key Id, value []byte) error {
	request := StoreRequest{
		PeerId: p.protocolNode.dhtNode.Peer.Id.Bytes(),
		Key: key.Bytes(),
		Value: value,
	}
	requestPayload, err := proto.Marshal(&request)
	if err != nil {
		return err
	}
	_, err = p.protocolNode.rpcNode.Call(
		p.addr, p.protocolNode.storeServiceId, requestPayload)
	return err
}
