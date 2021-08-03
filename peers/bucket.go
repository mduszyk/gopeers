package peers

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

func (b *bucket) remove(id Id) bool {
	i := b.find(id)
	if i > -1 {
		b.peers = append(b.peers[:i], b.peers[i+1:]...)
		return true
	}
	return false
}


type bucketList struct {
	buckets []bucket
}
