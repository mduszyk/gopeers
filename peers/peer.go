package peers

type Peer interface {
	Ping() error
	FindNode() error
}
