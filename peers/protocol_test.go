package peers

import (
	"testing"
)

func TestUdpProtocol(t *testing.T) {
	node1Peer, p2pNode1, err := NewUdpProtoNode(20, 5, "localhost:5001")
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node2Peer, p2pNode2, err := NewUdpProtoNode(20, 5, "localhost:5002")
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
	if i, _ := p2pNode1.buckets.find(p2pNode2.id); i < 0 {
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
	if i, _ := p2pNode1.buckets.find(p2pNode2.id); i < 0 {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}
