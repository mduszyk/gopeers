package peers

import (
	"math/big"
	"sort"
	"testing"
)

func TestTreeFind(t *testing.T) {
	tree := NewBucketTree(20)
	tree.split(tree.root)
	tree.split(tree.root.left)
	tree.split(tree.root.right)
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
	tree.split(tree.root)
	tree.split(tree.root.left)
	tree.split(tree.root.right)
	tree.split(tree.root.right.right)
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
		di := xor(peer.Id, allPeers[i].Id)
		dj := xor(peer.Id, allPeers[j].Id)
		return lt(di, dj)
	})
	expected := allPeers[:n]
	for i := 0; i < n; i++ {
		if !eq(expected[i].Id, peers[i].Id) {
			t.Errorf("ids don't match\n")
		}
	}
}

func TestTreeBuckets(t *testing.T) {
	tree := NewBucketTree(20)
	tree.split(tree.root)
	tree.split(tree.root.left)
	tree.split(tree.root.right)
	if tree.size != 4 {
		t.Errorf("invlid tree size\n")
	}
	buckets := tree.buckets(big.NewInt(0))
	if len(buckets) != tree.size {
		t.Errorf("invlid buckets count\n")
	}
	if buckets[0] != tree.root.left.left.bucket {
		t.Errorf("invlid bucket\n")
	}
	if buckets[1] != tree.root.left.right.bucket {
		t.Errorf("invlid bucket\n")
	}
	if buckets[2] != tree.root.right.left.bucket {
		t.Errorf("invlid bucket\n")
	}
	if buckets[3] != tree.root.right.right.bucket {
		t.Errorf("invlid bucket\n")
	}
	id := intBits([]uint{IdBits-1})
	buckets = tree.buckets(id)
	if len(buckets) != tree.size {
		t.Errorf("invlid buckets count\n")
	}
	if buckets[0] != tree.root.right.left.bucket {
		t.Errorf("invlid bucket\n")
	}
	if buckets[1] != tree.root.right.right.bucket {
		t.Errorf("invlid bucket\n")
	}
	if buckets[2] != tree.root.left.right.bucket {
		t.Errorf("invlid bucket\n")
	}
	if buckets[3] != tree.root.left.left.bucket {
		t.Errorf("invlid bucket\n")
	}
}