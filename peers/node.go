package peers

import (
	"errors"
	"time"
)

type p2pNode struct {
	b int
	peer *Peer
	tree *bucketTree
}

func NewP2pNode(k, b int, id Id) *p2pNode {
	node := &p2pNode{b: b, tree: NewBucketTree(k)}
	node.peer = &Peer{id, node, time.Now()}
	return node
}

func NewRandomIdP2pNode(k, b int) (*p2pNode, error) {
	nodeId, err := RandomId()
	if err != nil {
		return nil, err
	}
	node := NewP2pNode(k, b, nodeId)
	return node, nil
}

func (node *p2pNode) callPing(peer *Peer) error {
	randomId, err := RandomId()
	if err != nil {
		return err
	}
	echoId, err := peer.Proto.Ping(node.peer, randomId)
	if err != nil {
		return err
	}

	if !eq(echoId, randomId) {
		return errors.New("ping random id not echoed")
	}
	return nil
}

func (node *p2pNode) add(peer *Peer) bool {
	peer.touch()
	node.tree.mutex.Lock()
	n := node.tree.find(peer.Id)
	if n.bucket.isFull() {
		if n.bucket.inRange(node.peer.Id) || n.bucket.depth % node.b != 0 {
			node.tree.split(n)
			node.tree.mutex.Unlock()
			return node.add(peer)
		} else {
			j, leastSeenPeer := n.bucket.leastSeen()
			if j > -1 {
				err := node.callPing(leastSeenPeer)
				if err != nil {
					n.bucket.remove(j)
					node.tree.mutex.Unlock()
					return node.add(peer)
				} else {
					leastSeenPeer.touch()
				}
			}
			node.tree.mutex.Unlock()
			return false
		}
	} else {
		if j := n.bucket.find(peer.Id); j > -1 {
			n.bucket.remove(j)
		}
		added := n.bucket.add(peer)
		node.tree.mutex.Unlock()
		return added
	}
}

func (node *p2pNode) join(peer *Peer) error {
	node.add(peer)

	peers, err := peer.Proto.FindNode(node.peer, node.peer.Id)
	if err != nil {
		return err
	}
	for _, p := range peers {
		node.add(p)
	}

	node.tree.mutex.RLock()
	buckets := node.tree.buckets(peer.Id)
	node.tree.mutex.RUnlock()
	if len(buckets) > 1 {
		// skip bucket containing our id
		err = node.refreshBuckets(buckets[1:])
		if err != nil {
			return err
		}
	}

	return nil
}

func (node *p2pNode) refresh(b *bucket) error {
	id, err := RandomIdRange(b.lo, b.hi)
	if err != nil {
		return err
	}

	for _, peer := range b.peers {
		peers, err := peer.Proto.FindNode(node.peer, id)
		if err != nil {
			return err
		}
		for _, p := range peers {
			node.add(p)
		}
	}

	return nil
}

func (node *p2pNode) refreshBuckets(buckets []*bucket) error {
	for _, b := range buckets {
		err := node.refresh(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (node *p2pNode) refreshAll() error {
	node.tree.mutex.RLock()
	buckets := node.tree.buckets(node.peer.Id)
	node.tree.mutex.RUnlock()
	return node.refreshBuckets(buckets)
}

// Protocol interface

func (node *p2pNode) Ping(sender *Peer, randomId Id) (Id, error) {
	node.add(sender)
	return randomId, nil
}

func (node *p2pNode) FindNode(sender *Peer, id Id) ([]*Peer, error) {
	node.add(sender)
	node.tree.mutex.RLock()
	peers := node.tree.closest(id, node.tree.k)
	node.tree.mutex.RUnlock()
	return peers, nil
}
