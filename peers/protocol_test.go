package peers

import (
	"fmt"
	"log"
	"sync"
	"testing"
)

func TestUdpPing(t *testing.T) {
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

	randomId, err := CryptoRandId()
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
	if n := node1ProtoServer.p2pNode.Tree.Find(node2ProtoServer.p2pNode.Peer.Id); n.Bucket == nil {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}

func TestUdpJoin(t *testing.T) {
	n := 100
	k := 20
	b := 5
	basePort := 5000
	protoNodes := make([]*udpProtocolServer, n)
	peers := make([]*Peer, n)

	log.Printf("Generating nodes, n: %d", n)
	for i := 0; i < n; i++ {
		port := basePort + i
		protoNode, err := NewUdpProtoNode(k, b, fmt.Sprintf("localhost:%d", port))
		if err != nil {
			t.Errorf("failed creating node: %v\n", err)
		}
		protoNodes[i] = protoNode
		peer, err := NewRandomIdPeer()
		if err != nil {
			t.Errorf("failed creating peer: %v\n", err)
		}
		peers[i] = peer
	}

	log.Printf("Joining")
	var wg1 sync.WaitGroup
	for i := 1; i < n; i++ {
		protoNodes[i].Connect(protoNodes[0].rpcNode.Addr, peers[0])
		wg1.Add(1)
		go func(i int) {
			err := protoNodes[i].p2pNode.Join(peers[0])
			if err != nil {
				t.Errorf("failed joining: %v\n", err)
			}
			wg1.Done()
		}(i)
	}
	wg1.Wait()

	log.Printf("Refreshing")
	var wg2 sync.WaitGroup
	for i := 0; i < n; i++ {
		wg2.Add(1)
		go func(i int) {
			err := protoNodes[i].p2pNode.RefreshAll()
			if err != nil {
				t.Errorf("failed refreshing: %v\n", err)
			}
			wg2.Done()
		}(i)
	}
	wg2.Wait()

	log.Printf("Checking state")
	for i := 0; i < n; i++ {
		seen := 0
		for j := 0; j < n; j++ {
			if j == i {
				continue
			}
			node := protoNodes[i].p2pNode
			treeNode := node.Tree.Find(node.Peer.Id)
			if treeNode.Bucket.Contains(node.Peer.Id) {
				seen++
			}
		}
		log.Printf("seen: %d\n", seen)
		if seen < 1 {
			t.Errorf("node didn't join\n")
		}
	}
	log.Printf("Done")
}
