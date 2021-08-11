package peers

import (
	"errors"
	"time"
)

type p2pNode struct {
	b int
	peer *Peer
	bList *bucketList
}

func NewP2pNode(k, b int, id Id) *p2pNode {
	node := &p2pNode{
		b: b,
		bList: NewBucketList(k),
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

	if echoId != randomId {
		return errors.New("ping random id not echoed")
	}
	return nil
}

func (node *p2pNode) addPeer(peer *Peer) bool {
	peer.touch()
	i, bucket := node.bList.find(peer.Id)
	if bucket.isFull() {
		if bucket.inRange(node.peer.Id) || bucket.depth % node.b != 0 {
			node.bList.split(i)
			return node.addPeer(peer)
		} else {
			j, leastSeenPeer := bucket.leastSeen()
			err := node.pingPeer(leastSeenPeer)
			if err != nil {
				bucket.remove(j)
				return node.addPeer(peer)
			} else {
				leastSeenPeer.touch()
				return false
			}
		}
	} else {
		if j := bucket.find(peer.Id); j > -1 {
			bucket.remove(j)
		}
		return bucket.add(peer)
	}
}

// Protocol interface

func (node *p2pNode) Ping(sender *Peer, randomId Id) (Id, error) {
	node.addPeer(sender)
	return randomId, nil
}

func (node *p2pNode) FindNode(sender *Peer) error {
	return nil
}
