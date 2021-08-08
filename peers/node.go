package peers

type p2pNode struct {
	id Id
	buckets *bucketList
}

func NewP2pNode(id Id, buckets *bucketList) *p2pNode {
	node := &p2pNode{
		id: id,
		buckets: buckets,
	}
	return node
}

func (node *p2pNode) Ping(sender *Peer, randomId Id) (Id, error) {
	node.buckets.add(sender)
	return randomId, nil
}

func (node *p2pNode) FindNode(sender *Peer) error {
	return nil
}
