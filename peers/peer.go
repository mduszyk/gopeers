package peers

import "time"

type Peer struct {
	Id Id
	Protocol Protocol
	LastSeen time.Time
}

func (p *Peer) touch() {
	p.LastSeen = time.Now()
}