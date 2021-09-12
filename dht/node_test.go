package dht

import (
	"fmt"
	"github.com/mduszyk/gopeers/store"
	"log"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestAddFind(t *testing.T) {
	nodeId := big.NewInt(0)
	k := 20
	b := 5
	alpha := 3
	storage := store.NewMemStorage()
	node := NewKadNode(k, b, alpha, nodeId, storage)
	for i := 0; i < k; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		peer := &Peer{id, nil, time.Now()}
		if !node.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
	}
	for i := 0; i < k; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		n := node.Tree.Find(id)
		if !n.Bucket.Contains(id) {
			t.Errorf("bucket should contain given Id\n")
		}
	}
}

func TestBucketListSplit(t *testing.T) {
	nodeId := big.NewInt(0)
	k := 20
	b := 5
	alpha := 3
	storage := store.NewMemStorage()
	node := NewKadNode(k, b, alpha, nodeId, storage)
	for i := 0; i < k+1; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		peer := &Peer{id, nil, time.Now()}
		if !node.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
	}
	if node.Tree.root.left == nil || node.Tree.root.right == nil {
		t.Errorf("there should be 2 buckets\n")
	}
	for i := 0; i < k+1; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		n := node.Tree.Find(id)
		if !n.Bucket.Contains(id) {
			t.Errorf("bucket should contain given Id\n")
		}
	}
}

func TestNodePing(t *testing.T) {
	storage := store.NewMemStorage()
	p2pNode1, err := NewRandomIdKadNode(20, 5, 3, storage)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	p2pNode2, err := NewRandomIdKadNode(20, 5, 3, storage)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	randomId, err := CryptoRandId()
	if err != nil {
		t.Errorf("failed generating random Id: %v\n", err)
	}
	echoId, err := p2pNode1.Ping(p2pNode2.Peer, randomId)
	if err != nil {
		t.Errorf("failed pinging: %v\n", err)
	}
	if !eq(echoId, randomId) {
		t.Errorf("ping returned invalid Id\n")
	}
	if n := p2pNode1.Tree.Find(p2pNode2.Peer.Id); n.Bucket == nil {
		t.Errorf("id of node 2 not added to bucket in node 1\n")
	}
}

func TestNodeTrivialJoin(t *testing.T) {
	storage := store.NewMemStorage()
	node1, err := NewRandomIdKadNode(20, 5, 3, storage)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	node2, err := NewRandomIdKadNode(20, 5, 3, storage)
	if err != nil {
		t.Errorf("failed creating node: %v\n", err)
	}

	err = node1.Join(node2.Peer)
	if err != nil {
		t.Errorf("failed joining: %v\n", err)
	}
	if n := node2.Tree.Find(node1.Peer.Id); !n.Bucket.Contains(node1.Peer.Id) {
		t.Errorf("id not added to bucket\n")
	}
	if n := node1.Tree.Find(node2.Peer.Id); !n.Bucket.Contains(node2.Peer.Id) {
		t.Errorf("id not added to bucket\n")
	}
}

func TestNodeJoin(t *testing.T) {
	n := 400
	k := 20
	b := 5
	alpha := 3
	nodes := make([]*KadNode, n)

	log.Printf("Generating nodes, n: %d", n)
	for i := 0; i < n; i++ {
		storage := store.NewMemStorage()
		node, err := NewRandomIdKadNode(k, b, alpha, storage)
		if err != nil {
			t.Errorf("failed creating node: %v\n", err)
		}
		nodes[i] = node
	}

	log.Printf("Joining")
	var wg1 sync.WaitGroup
	for i := 1; i < n; i++ {
		wg1.Add(1)
		go func(i int) {
			err := nodes[i].Join(nodes[0].Peer)
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
			err := nodes[i].RefreshAll()
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
			treeNode := nodes[j].Tree.Find(nodes[i].Peer.Id)
			if treeNode.Bucket.Contains(nodes[i].Peer.Id) {
				seen++
			}
		}
		if seen < k {
			t.Errorf("node didn't join\n")
		}
	}
	log.Printf("Done")
}

func TestLookupSetGet(t *testing.T) {
	n := 400
	k := 20
	b := 5
	alpha := 3
	nodes := make([]*KadNode, n)
	nodePeers := make([]*Peer, 0, n)

	log.Printf("Generating nodes, n: %d", n)
	for i := 0; i < n; i++ {
		storage := store.NewMemStorage()
		node, err := NewRandomIdKadNode(k, b, alpha, storage)
		nodePeers = append(nodePeers, node.Peer)
		if err != nil {
			t.Errorf("failed creating node: %v\n", err)
		}
		nodes[i] = node
	}

	log.Printf("Joining")
	var wg1 sync.WaitGroup
	for i := 1; i < n; i++ {
		wg1.Add(1)
		go func(i int) {
			err := nodes[i].Join(nodes[0].Peer)
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
			err := nodes[i].RefreshAll()
			if err != nil {
				t.Errorf("failed refreshing: %v\n", err)
			}
			wg2.Done()
		}(i)
	}
	wg2.Wait()

	log.Printf("Lookup")
	id := MathRandId()
	peers := nodes[0].Lookup(id)
	sortByDistance(nodePeers, id)
	expectedPeers := nodePeers[:k]
	if len(peers) != len(expectedPeers) {
		t.Errorf("lookup returned wrong number of peers\n")
	}
	for i, peer := range peers {
		if !eq(peer.Id, expectedPeers[i].Id) {
			t.Errorf("unexpected peer: %d\n", i)
		}
	}

	key := MathRandId().Bytes()
	value := []byte("test")
	err := nodes[0].Set(key, value)
	if err != nil {
		t.Errorf("set failed: %v\n", err)
	}

	value2, err := nodes[n-1].Get(key)
	if err != nil {
		t.Errorf("get failed: %v\n", err)
	}
	if !reflect.DeepEqual(value2, value) {
		t.Errorf("got invalid value: %v\n", value)
	}

	log.Printf("Done")
}
