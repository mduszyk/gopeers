package peers

import (
	"errors"
	"time"
)

type p2pNode struct {
	b int
	peer *Peer
	buckets *bucketTree
}

func NewP2pNode(k, b int, id Id) *p2pNode {
	node := &p2pNode{
		b: b,
		buckets: NewBucketTree(k),
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

func (node *p2pNode) addPeer(peer *Peer) bool {
	peer.touch()
	n := node.buckets.find(peer.Id)
	if n.bucket.isFull() {
		if n.bucket.inRange(node.peer.Id) || n.bucket.depth % node.b != 0 {
			n.split()
			return node.addPeer(peer)
		} else {
			j, leastSeenPeer := n.bucket.leastSeen()
			err := node.pingPeer(leastSeenPeer)
			if err != nil {
				n.bucket.remove(j)
				return node.addPeer(peer)
			} else {
				leastSeenPeer.touch()
				return false
			}
		}
	} else {
		if j := n.bucket.find(peer.Id); j > -1 {
			n.bucket.remove(j)
		}
		return n.bucket.add(peer)
	}
}

// Protocol interface

func (node *p2pNode) Ping(sender *Peer, randomId Id) (Id, error) {
	node.addPeer(sender)
	return randomId, nil
}

func (node *p2pNode) FindNode(sender *Peer, id Id) ([]*Peer, error) {
	peers := node.buckets.closest(id, node.buckets.k)
	return peers, nil
}
