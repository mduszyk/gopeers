package peers

import "time"

type Peer struct {
	Id Id
	Protocol Protocol
	LastSeen time.Time
}