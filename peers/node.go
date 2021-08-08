package peers

import "errors"

type p2pNode struct {
	b int
	peer *Peer
	bList *bucketList
}

func NewP2pNode(k, b int, peer *Peer) *p2pNode {
	node := &p2pNode{
		b: b,
		peer: peer,
		bList: NewBucketList(k),
	}
	return node
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
	i, b := node.bList.find(peer.Id)
	if b.isFull() {
		if b.inRange(node.peer.Id) || b.depth() % node.b != 0 {
			node.bList.split(i)
			return node.addPeer(peer)
		} else {
			j, leastSeenPeer := b.leastSeen()
			err := node.pingPeer(leastSeenPeer)
			if err != nil {
				b.remove(j)
				return node.addPeer(peer)
			} else {
				leastSeenPeer.touch()
				return false
			}
		}
	} else {
		if j := b.find(peer.Id); j > -1 {
			b.remove(j)
		}
		return b.add(peer)
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
