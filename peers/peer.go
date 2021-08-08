package peers

import "time"

type Peer struct {
	Id       Id
	Proto    Protocol
	LastSeen time.Time
}

func (p *Peer) touch() {
	p.LastSeen = time.Now()
}