package peers

import (
	"math/big"
	"sort"
)

type bucket struct {
	k, depth int
	lo, hi Id
	peers []*Peer
}

func NewBucket(k, depth int, lo Id, hi Id) *bucket {
	peers := make([]*Peer, 0, k)
	return &bucket{k:k, depth: depth, lo: lo, hi: hi, peers: peers}
}

func (b *bucket) inRange(id Id) bool {
	// lo <= id < hi
	return lte(b.lo, id) && lt(id, b.hi)
}

func (b *bucket) isFull() bool {
	return len(b.peers) >= b.k
}

func (b *bucket) find(id Id) int {
   for i, peer := range b.peers {
	   if eq(id, peer.Id) {
		   return i
	   }
	}
	return -1
}

func (b *bucket) Contains(id Id) bool {
	return b.find(id) > -1
}

func (b *bucket) add(peer *Peer) bool {
	if !b.isFull() {
		b.peers = append(b.peers, peer)
		return true
	}
	return false
}

func (b *bucket) remove(i int) bool {
	if i > -1 {
		b.peers = append(b.peers[:i], b.peers[i+1:]...)
		return true
	}
	return false
}

func (b *bucket) split() (*bucket, *bucket) {
	middle := new(big.Int).Div(new(big.Int).Add(b.lo, b.hi), big.NewInt(2))
	b1 := NewBucket(b.k, b.depth + 1, b.lo, middle)
	b2 := NewBucket(b.k, b.depth + 1, middle, b.hi)
	for _, peer := range b.peers {
		if b1.inRange(peer.Id) {
			b1.add(peer)
		} else {
			b2.add(peer)
		}
	}
	return b1, b2
}

func (b *bucket) leastSeen() (int, *Peer) {
	if len(b.peers) > 0 {
		return 0, b.peers[0]
	}
	return -1, nil
}

func (b *bucket) closest(id Id, n int) []*Peer {
	peers := make([]*Peer, len(b.peers))
	copy(peers, b.peers)
	sort.Slice(peers, func(i, j int) bool {
		di := xor(id, peers[i].Id)
		dj := xor(id, peers[j].Id)
		return lt(di, dj)
	})
	return peers[:min(n, len(peers))]
}
