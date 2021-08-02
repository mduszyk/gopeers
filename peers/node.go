package peers

type p2pNode struct {
	id Id
	buckets []kBucket
}

func NewP2pNode() (*p2pNode, error) {
	id, err := RandomId()
	if err != nil {
		return nil, err
	}
	node := &p2pNode{
		id: id,
	}
	return node, nil
}

func (node *p2pNode) Ping(sender Peer) error {
	return nil
}

func (node *p2pNode) FindNode(sender Peer) error {
	return nil
}
