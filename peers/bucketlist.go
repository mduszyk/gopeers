package peers

import "math/big"

type bucketList struct {
	k, splitLevel int
	nodePeer      *Peer
	buckets       []*bucket
}

func NewBucketList(k int, splitLevel int, node *Peer) *bucketList {
	b := NewBucket(k, big.NewInt(0), maxId)
	buckets := make([]*bucket, 0, k)
	buckets = append(buckets, b)
	return &bucketList{k, splitLevel, node, buckets}
}

func (bl *bucketList) find(id Id) (int, *bucket) {
	for i, b := range bl.buckets {
		if b.inRange(id) {
			return i, b
		}
	}
	return -1, nil
}

func (bl *bucketList) add(peer *Peer) bool {
	peer.touch()
	i, b := bl.find(peer.Id)
	if b.isFull() {
		if b.inRange(bl.nodePeer.Id) || b.depth() % bl.splitLevel != 0 {
			bl.split(i, b)
			return bl.add(peer)
		} else {
			if bl.pingLeastSeen(b) {
				return false
			} else {
				return bl.add(peer)
			}
		}
	} else if j := b.find(peer.Id); j > -1 {
		b.remove(j)
		return b.add(peer)
	} else {
		return b.add(peer)
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

func (bl *bucketList) pingLeastSeen(b *bucket) bool {
	i, peer := b.leastSeen()
	randomId, err := RandomId()
	if err != nil {
		b.remove(i)
		return false
	}
	id, err := peer.Proto.Ping(bl.nodePeer, randomId)
	if id != randomId || err != nil {
		b.remove(i)
		return false
	} else {
		peer.touch()
		return true
	}
}
