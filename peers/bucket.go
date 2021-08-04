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

type bucketList struct {
	k, splitLevel int
	nodeId        Id
	buckets []*bucket
}

func NewBucketList(k int, splitLevel int, nodeId Id) *bucketList {
	b := NewBucket(k, big.NewInt(0), maxId)
	buckets := make([]*bucket, 0, k)
	buckets = append(buckets, b)
	return &bucketList{k, splitLevel, nodeId, buckets}
}

func (bl *bucketList) find(id Id) *bucket {
	for _, b := range bl.buckets {
		if b.inRange(id) {
			return b
		}
	}
	return nil
}

func (bl *bucketList) add(peer Peer) {
	peer.touch()
	b := bl.find(peer.Id)
	if b.isFull() {
		if b.inRange(bl.nodeId) || b.depth() % bl.splitLevel != 0 {
			bl.split(b, peer)
		} else {
			bl.pingLastSeen(b, peer)
		}
	} else if i := b.find(peer.Id); i > -1 {
		b.remove(i)
		b.add(peer)
	} else {
		b.add(peer)
	}
}

func (bl *bucketList) split(bk *bucket, peer Peer) {

}

func (bl *bucketList) pingLastSeen(bk *bucket, peer Peer) {

}