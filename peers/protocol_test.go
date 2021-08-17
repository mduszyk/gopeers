package peers

import (
	"sync"
	"testing"
)

func TestUdpProtocol(t *testing.T) {
	node1ProtoServer, err := NewUdpProtoNode(20, 5, "localhost:5001")
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node2ProtoServer, err := NewUdpProtoNode(20, 5, "localhost:5002")
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node1Peer, err := NewRandomIdPeer()
	if err != nil {
		t.Errorf("failed creating peer: %v\n", err)
	}
	node2Peer, err := NewRandomIdPeer()
	if err != nil {
		t.Errorf("failed creating peer: %v\n", err)
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
	if !eq(echoId, randomId) {
		t.Errorf("ping returned invalid Id\n")
	}
	if n := node1ProtoServer.p2pNode.tree.find(node2ProtoServer.p2pNode.peer.Id); n.bucket == nil {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}

func TestMethodCallProtocol(t *testing.T) {
	p2pNode1, err := NewRandomIdP2pNode(20, 5)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	p2pNode2, err := NewRandomIdP2pNode(20, 5)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	randomId, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random Id: %v\n", err)
	}
	echoId, err := p2pNode1.Ping(p2pNode2.peer, randomId)
	if err != nil {
		t.Errorf("failed pinging: %v\n", err)
	}
	if !eq(echoId, randomId) {
		t.Errorf("ping returned invalid Id\n")
	}
	if n := p2pNode1.tree.find(p2pNode2.peer.Id); n.bucket == nil {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}

func TestMethodCallTrivialJoin(t *testing.T) {
	node1, err := NewRandomIdP2pNode(20, 5)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node2, err := NewRandomIdP2pNode(20, 5)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	err = node1.join(node2.peer)
	if err != nil {
		t.Errorf("failed joining: %v\n", err)
	}
	if n := node2.tree.find(node1.peer.Id); !n.bucket.contains(node1.peer.Id) {
		t.Errorf("id not added to bucket\n")
	}
	if n := node1.tree.find(node2.peer.Id); !n.bucket.contains(node2.peer.Id) {
		t.Errorf("id not added to bucket\n")
	}
}

func TestMethodCallJoin(t *testing.T) {
	n := 200
	k := 20
	b := 5
	nodes := make([]*p2pNode, n)
	for i := 0; i < n; i++ {
		node, err := NewRandomIdP2pNode(k, b)
		if err != nil {
			t.Errorf("failed creating node: %v\n", err)
		}
		nodes[i] = node
	}

	var wg1 sync.WaitGroup
	for i := 1; i < n; i++ {
		wg1.Add(1)
		go func(i int) {
			err := nodes[i].join(nodes[0].peer)
			if err != nil {
				t.Errorf("failed joining: %v\n", err)
			}
			wg1.Done()
		}(i)
	}
	wg1.Wait()

	var wg2 sync.WaitGroup
	for i := 0; i < n; i++ {
		wg2.Add(1)
		go func(i int) {
			err := nodes[i].refreshAll()
			if err != nil {
				t.Errorf("failed refreshing: %v\n", err)
			}
			wg2.Done()
		}(i)
	}
	wg2.Wait()

	for i := 0; i < n; i++ {
		seen := 0
		for j := 0; j < n; j++ {
			if j == i {
				continue
			}
			treeNode := nodes[j].tree.find(nodes[i].peer.Id)
			if treeNode.bucket.contains(nodes[i].peer.Id) {
				seen++
			}
		}
		if seen < k {
			t.Errorf("node didn't join\n")
		}
	}
}
