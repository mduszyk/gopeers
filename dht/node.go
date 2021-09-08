package dht

import (
	"errors"
	"time"
)

type KadNode struct {
	b    int
	Peer *Peer
	Tree *bucketTree
}

func NewKadNode(k, b int, id Id) *KadNode {
	node := &KadNode{b: b, Tree: NewBucketTree(k)}
	node.Peer = &Peer{id, node, time.Now()}
	return node
}

func NewRandomIdKadNode(k, b int) (*KadNode, error) {
	nodeId, err := CryptoRandId()
	if err != nil {
		return nil, err
	}
	node := NewKadNode(k, b, nodeId)
	return node, nil
}

func (node *KadNode) callPing(peer *Peer) error {
	randomId, err := CryptoRandId()
	if err != nil {
		return err
	}
	echoId, err := peer.Proto.Ping(node.Peer, randomId)
	if err != nil {
		return err
	}

	if !eq(echoId, randomId) {
		return errors.New("ping random id not echoed")
	}
	return nil
}

func (node *KadNode) add(peer *Peer) bool {
	peer.touch()
	node.Tree.mutex.Lock()
	n := node.Tree.Find(peer.Id)
	if i := n.Bucket.find(peer.Id); i > -1 {
		n.Bucket.remove(i)
		added := n.Bucket.add(peer)
		node.Tree.mutex.Unlock()
		return added
	} else if n.Bucket.isFull() {
		if n.Bucket.inRange(node.Peer.Id) || n.Bucket.depth % node.b != 0 {
			node.Tree.split(n)
			node.Tree.mutex.Unlock()
			return node.add(peer)
		} else {
			if j, leastSeenPeer := n.Bucket.leastSeen(); j > -1 {
				node.Tree.mutex.Unlock()
				err := node.callPing(leastSeenPeer)
				node.Tree.mutex.Lock()
				if k := n.Bucket.find(leastSeenPeer.Id); k > -1 {
					n.Bucket.remove(k)
				}
				if err != nil {
					node.Tree.mutex.Unlock()
					return node.add(peer)
				} else {
					leastSeenPeer.touch()
					n.Bucket.add(leastSeenPeer)
				}
			}
			node.Tree.mutex.Unlock()
			return false
		}
	} else {
		added := n.Bucket.add(peer)
		node.Tree.mutex.Unlock()
		return added
	}
}

func (node *KadNode) Join(peer *Peer) error {
	node.add(peer)

	peers, err := peer.Proto.FindNode(node.Peer, node.Peer.Id)
	if err != nil {
		return err
	}
	for _, p := range peers {
		node.add(p)
	}

	node.Tree.mutex.RLock()
	buckets := node.Tree.buckets(peer.Id)
	node.Tree.mutex.RUnlock()
	if len(buckets) > 1 {
		// skip bucket containing our id
		err = node.refreshBuckets(buckets[1:])
		if err != nil {
			return err
		}
	}

	return nil
}

func (node *KadNode) refresh(b *bucket) error {
	//id, err := CryptoRandIdRange(b.lo, b.hi)
	//if err != nil {
	//	return err
	//}
	id := MathRandIdRange(b.lo, b.hi)

	for _, peer := range b.peers {
		peers, err := peer.Proto.FindNode(node.Peer, id)
		if err != nil {
			return err
		}
		for _, p := range peers {
			node.add(p)
		}
	}

	return nil
}

func (node *KadNode) refreshBuckets(buckets []*bucket) error {
	for _, b := range buckets {
		err := node.refresh(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (node *KadNode) RefreshAll() error {
	node.Tree.mutex.RLock()
	buckets := node.Tree.buckets(node.Peer.Id)
	node.Tree.mutex.RUnlock()
	return node.refreshBuckets(buckets)
}

func (node *KadNode) Lookup(id Id) []*Peer {
	// TODO
	return nil
}

// Protocol interface

func (node *KadNode) Ping(sender *Peer, randomId Id) (Id, error) {
	node.add(sender)
	return randomId, nil
}

func (node *KadNode) FindNode(sender *Peer, id Id) ([]*Peer, error) {
	node.add(sender)
	node.Tree.mutex.RLock()
	peers := node.Tree.closest(id, node.Tree.k)
	node.Tree.mutex.RUnlock()
	return peers, nil
}

func (node *KadNode) FindValue(sender *Peer, key Id) (*FindResult, error) {
	// TODO
	return nil, nil
}

func (node *KadNode) Store(sender *Peer, key Id, value []byte) error {
	// TODO
	return nil
}
