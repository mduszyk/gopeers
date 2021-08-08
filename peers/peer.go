package peers

import "time"

type Peer struct {
	Id       Id
	Proto    Protocol
	LastSeen time.Time
}

func NewRandomIdPeer() (*Peer, error){
	id, err := RandomId()
	if err != nil {
		return nil, err
	}
	peer := &Peer{Id: id, LastSeen: time.Now()}
	return peer, nil
}

func (p *Peer) touch() {
	p.LastSeen = time.Now()
}