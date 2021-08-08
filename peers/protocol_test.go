package peers

import (
	"testing"
)

func TestUdpProtocol(t *testing.T) {
	node1Peer, node1ProtoServer, err := NewUdpProtoNode(20, 5, "localhost:5001")
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node2Peer, node2ProtoServer, err := NewUdpProtoNode(20, 5, "localhost:5002")
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node1ProtoServer.Connect(node2ProtoServer.rpcNode.Addr, node2Peer)
	node2ProtoServer.Connect(node1ProtoServer.rpcNode.Addr, node1Peer)

	randomId, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random Id: %v\n", err)
	}
	echoId, err := node1Peer.Proto.Ping(node2Peer, randomId)
	if err != nil {
		t.Errorf("failed pinging: %v\n", err)
	}
	if echoId.Cmp(randomId) != 0 {
		t.Errorf("ping returned invalid Id\n")
	}
	if i, _ := node1ProtoServer.p2pNode.bList.find(node2ProtoServer.p2pNode.peer.Id); i < 0 {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}

func TestMethodCallProtocol(t *testing.T) {
	node1Peer, p2pNode1, err := NewMethodCallProtoNode(20, 5)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node2Peer, p2pNode2, err := NewMethodCallProtoNode(20, 5)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	randomId, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random Id: %v\n", err)
	}
	echoId, err := node1Peer.Proto.Ping(node2Peer, randomId)
	if err != nil {
		t.Errorf("failed pinging: %v\n", err)
	}
	if echoId.Cmp(randomId) != 0 {
		t.Errorf("ping returned invalid Id\n")
	}
	if i, _ := p2pNode1.bList.find(p2pNode2.peer.Id); i < 0 {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}
