package peers

type Protocol interface {
	Ping(sender Peer) error
	FindNode(sender Peer) error
}
