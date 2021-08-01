package peers

type p2pNode struct {

}

func NewP2pNode() (*p2pNode, error) {
	node := &p2pNode{}
	return node, nil
}

func (node *p2pNode) Ping() error {
	return nil
}

func (node *p2pNode) FindNode() error {
	return nil
}
