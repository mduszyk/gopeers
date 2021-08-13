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
	node := &p2pNode{
		b: b,
		tree: NewBucketTree(k),
	}
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

func (node *p2pNode) pingPeer(peer *Peer) error {
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
	n := node.tree.find(peer.Id)
	if n.bucket.isFull() {
		if n.bucket.inRange(node.peer.Id) || n.bucket.depth % node.b != 0 {
			node.tree.split(n)
			return node.add(peer)
		} else {
			j, leastSeenPeer := n.bucket.leastSeen()
			if j > -1 {
				err := node.pingPeer(leastSeenPeer)
				if err != nil {
					n.bucket.remove(j)
					return node.add(peer)
				} else {
					leastSeenPeer.touch()
				}
			}
			return false
		}
	} else {
		if j := n.bucket.find(peer.Id); j > -1 {
			n.bucket.remove(j)
		}
		return n.bucket.add(peer)
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

	buckets := node.tree.buckets(peer.Id)
	if len(buckets) > 1 {
		// skip bucket containing peer.Id
		for _, b := range buckets[1:] {
			err = node.refresh(b)
			if err != nil {
				return err
			}
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

// Protocol interface

func (node *p2pNode) Ping(sender *Peer, randomId Id) (Id, error) {
	node.add(sender)
	return randomId, nil
}

func (node *p2pNode) FindNode(sender *Peer, id Id) ([]*Peer, error) {
	node.add(sender)
	peers := node.tree.closest(id, node.tree.k)
	return peers, nil
}
