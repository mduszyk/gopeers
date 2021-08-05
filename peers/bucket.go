package peers

import "math/big"

type bucket struct {
	k int
	lo Id
	hi Id
	peers []Peer
}

func NewBucket(k int, lo Id, hi Id) *bucket {
	peers := make([]Peer, 0, k)
	return &bucket{k, lo, hi, peers}
}

// lo <= id < hi
func (b *bucket) inRange(id Id) bool {
	r := b.lo.Cmp(id)
	return (r == 0 || r == -1) && b.hi.Cmp(id) == 1
}

func (b *bucket) isFull() bool {
	return len(b.peers) >= b.k
}

func (b *bucket) find(id Id) int {
   for i, peer := range b.peers {
	   if id.Cmp(peer.Id) == 0 {
		   return i
	   }
	}
	return -1
}

func (b *bucket) contains(id Id) bool {
	return b.find(id) > -1
}

func (b *bucket) add(peer Peer) bool {
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

func (b *bucket) depth() int {
	if len(b.peers) == 0 {
		return 0
	}
	prefix := ToBits(b.peers[0].Id)
	for _, peer := range b.peers[1:] {
		prefix = SharedBits(prefix, peer.Id)
	}
	return len(prefix)
}

func (b *bucket) split() (*bucket, *bucket) {
	middle := new(big.Int).Div(new(big.Int).Add(b.lo, b.hi), big.NewInt(2))
	b1 := NewBucket(b.k, b.lo, middle)
	b2 := NewBucket(b.k, middle, b.hi)
	for _, peer := range b.peers {
		if b1.inRange(peer.Id) {
			b1.add(peer)
		} else {
			b2.add(peer)
		}
	}
	return b1, b2
}

func (b *bucket) leastSeen() (int, Peer) {
	peer := b.peers[0]
	index := -1
	for i, p := range b.peers[1:] {
		if p.LastSeen.Before(peer.LastSeen) {
			peer = p
			index = i
		}
	}
	return index, peer
}
