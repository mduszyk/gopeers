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

type bucketList struct {
	k, splitLevel int
	nodeId Id
	nodePeer Peer
	buckets []*bucket
}

func NewBucketList(k int, splitLevel int, nodeId Id, nodePeer Peer) *bucketList {
	b := NewBucket(k, big.NewInt(0), maxId)
	buckets := make([]*bucket, 0, k)
	buckets = append(buckets, b)
	return &bucketList{k, splitLevel,
		nodeId, nodePeer, buckets}
}

func (bl *bucketList) find(id Id) (int, *bucket) {
	for i, b := range bl.buckets {
		if b.inRange(id) {
			return i, b
		}
	}
	return -1, nil
}

func (bl *bucketList) add(peer Peer) {
	peer.touch()
	i, b := bl.find(peer.Id)
	if b.isFull() {
		if b.inRange(bl.nodeId) || b.depth() % bl.splitLevel != 0 {
			bl.split(i, b)
			bl.add(peer)
		} else {
			bl.pingLeastSeen(b)
		}
	} else if j := b.find(peer.Id); j > -1 {
		b.remove(j)
		b.add(peer)
	} else {
		b.add(peer)
	}
}

func insert(a []*bucket, i int, value *bucket) []*bucket {
    if len(a) == i { // nil or empty slice or after last element
        return append(a, value)
    }
    a = append(a[:i+1], a[i:]...) // index < len(a)
    a[i] = value
    return a
}

func (bl *bucketList) split(i int, b *bucket) {
	bucket1, bucket2 := b.split()
	bl.buckets[i] = bucket1
	bl.buckets = insert(bl.buckets, i + 1, bucket2)
}

func (bl *bucketList) pingLeastSeen(b *bucket) {
	i, peer := b.leastSeen()
	err := peer.Protocol.Ping(bl.nodePeer)
	if err != nil {
		b.remove(i)
	} else {
		peer.touch()
	}
}