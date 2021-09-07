package dht

import (
	"fmt"
	"log"
	"math/big"
	"sync"
	"testing"
	"time"
)

var callTimeout = time.Minute
var bufferSize = uint32(10240)

func TestUdpPing(t *testing.T) {
	node1ProtoServer, err := NewUdpProtocolNode(20, 5, "localhost:4001", callTimeout, bufferSize)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node2ProtoServer, err := NewUdpProtocolNode(20, 5, "localhost:4002", callTimeout, bufferSize)
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
	if n := node1ProtoServer.dhtNode.Tree.Find(node2ProtoServer.dhtNode.Peer.Id); n.Bucket == nil {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}

func TestUdpFindNode(t *testing.T) {
	node1, err := NewUdpProtocolNode(20, 5, "localhost:5001", callTimeout, bufferSize)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node2, err := NewUdpProtocolNode(20, 5, "localhost:5002", callTimeout, bufferSize)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node1.dhtNode.add(node2.dhtNode.Peer)

	node1Peer := NewPeer(node1.dhtNode.Peer.Id)
	node2.Connect(node1.rpcNode.Addr, node1Peer)

	id := big.NewInt(0)
	// node2 calls node1
	peers, err := node1Peer.Proto.FindNode(node2.dhtNode.Peer, id)
	if err != nil {
		t.Errorf("failed finding nodes: %v\n", err)
	}
	if len(peers) != 1 {
		t.Errorf("got incorrect number of nodes\n")
	}
	if !eq(peers[0].Id, node2.dhtNode.Peer.Id) {
		t.Errorf("found incorrect peer\n")
	}
}

func TestUdpJoin(t *testing.T) {
	n := 100
	k := 20
	b := 5
	basePort := 6000
	protoNodes := make([]*udpProtocolNode, n)
	peers := make([]*Peer, n)

	log.Printf("Generating nodes, n: %d", n)
	for i := 0; i < n; i++ {
		port := basePort + i
		address := fmt.Sprintf("localhost:%d", port)
		protoNode, err := NewUdpProtocolNode(k, b, address, callTimeout, bufferSize)
		if err != nil {
			t.Errorf("failed creating node: %v\n", err)
		}
		protoNodes[i] = protoNode
		peers[i] = NewPeer(protoNodes[0].dhtNode.Peer.Id)
	}

	log.Printf("Joining")
	var wg1 sync.WaitGroup
	for i := 1; i < n; i++ {
		protoNodes[i].Connect(protoNodes[0].rpcNode.Addr, peers[i])
		wg1.Add(1)
		go func(i int) {
			err := protoNodes[i].dhtNode.Join(peers[i])
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
			err := protoNodes[i].dhtNode.RefreshAll()
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
			node := protoNodes[i].dhtNode
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
