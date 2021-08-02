package peers

type kBucket struct {
	lo Id
	hi Id
	peers []*Peer
}
