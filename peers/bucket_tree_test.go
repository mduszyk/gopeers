package peers

import (
	"math/big"
	"sort"
	"testing"
)

func TestTreeFind(t *testing.T) {
	tree := NewBucketTree(20)
	tree.root.split()
	tree.root.left.split()
	tree.root.right.split()
	peer := &Peer{Id: intBits([]uint{IdBits - 1, IdBits - 2})}
	n := tree.find(peer.Id)
	if n != tree.root.right.right {
		t.Errorf("found wrong node for id: %0160b\n", peer.Id)
	}
	peer = &Peer{Id: intBits([]uint{IdBits - 1})}
	n = tree.find(peer.Id)
	if n != tree.root.right.left {
		t.Errorf("found wrong node for id: %0160b\n", peer.Id)
	}
	peer = &Peer{Id: intBits([]uint{IdBits - 2})}
	n = tree.find(peer.Id)
	if n != tree.root.left.right {
		t.Errorf("found wrong node for id: %0160b\n", peer.Id)
	}
	peer = &Peer{Id: intBits([]uint{0})}
	n = tree.find(peer.Id)
	if n != tree.root.left.left {
		t.Errorf("found wrong node for id: %0160b\n", peer.Id)
	}
}

func TestTreeClosest(t *testing.T) {
	tree := NewBucketTree(20)
	tree.root.split()
	tree.root.left.split()
	tree.root.right.split()
	tree.root.right.right.split()
	allPeers := make([]*Peer, 0, 100)
	setBits1 := []uint{IdBits - 1, IdBits - 2, IdBits - 3}
	bucket1 := tree.root.right.right.right.bucket
	for i := 0; i < 10; i++ {
		setBits1 = append(setBits1, uint(i))
		peer := &Peer{Id: intBits(setBits1)}
		allPeers = append(allPeers, peer)
		if !bucket1.add(peer) {
			t.Errorf("failed adding to bucket\n")
		}
	}
	setBits2 := []uint{IdBits - 1, IdBits - 2}
	bucket2 := tree.root.right.right.left.bucket
	for i := 0; i < 5; i++ {
		setBits2 = append(setBits2, uint(i))
		peer := &Peer{Id: intBits(setBits2)}
		allPeers = append(allPeers, peer)
		if !bucket2.add(peer) {
			t.Errorf("failed adding to bucket\n")
		}
	}
	setBits3 := []uint{IdBits - 1}
	bucket3 := tree.root.right.left.bucket
	for i := 0; i < 5; i++ {
		setBits3 = append(setBits3, uint(i))
		peer := &Peer{Id: intBits(setBits3)}
		allPeers = append(allPeers, peer)
		if !bucket3.add(peer) {
			t.Errorf("failed adding to bucket\n")
		}
	}
	setBits4 := []uint{32}
	bucket4 := tree.root.left.left.bucket
	for i := 0; i < 5; i++ {
		setBits4 = append(setBits4, uint(i))
		peer := &Peer{Id: intBits(setBits4)}
		allPeers = append(allPeers, peer)
		if !bucket4.add(peer) {
			t.Errorf("failed adding to bucket\n")
		}
	}
	setBits := []uint{IdBits - 1, IdBits - 2, IdBits - 3, 10}
	peer := &Peer{Id: intBits(setBits)}
	n := 18
	peers := tree.closest(peer.Id, n)
	if len(peers) != n {
		t.Errorf("returned wrong number of close peers: %d\n", len(peers))
	}
	sort.Slice(allPeers, func(i, j int) bool {
		di := new(big.Int).Xor(peer.Id, allPeers[i].Id)
		dj := new(big.Int).Xor(peer.Id, allPeers[j].Id)
		return di.Cmp(dj) == -1
	})
	expected := allPeers[:n]
	for i := 0; i < n; i++ {
		if expected[i].Id.Cmp(peers[i].Id) != 0 {
			t.Errorf("ids don't match\n")
		}
	}
}