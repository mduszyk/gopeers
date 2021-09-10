package dht

import (
	"sort"
	"time"
)

type Peer struct {
	Id       Id
	Proto    Protocol
	LastSeen time.Time
}

func NewPeer(id Id) *Peer {
	return &Peer{Id: id, LastSeen: time.Now()}
}

func NewRandomIdPeer() (*Peer, error){
	id, err := CryptoRandId()
	if err != nil {
		return nil, err
	}
	peer := &Peer{Id: id, LastSeen: time.Now()}
	return peer, nil
}

func (p *Peer) touch() {
	p.LastSeen = time.Now()
}

func sortByDistance(peers []*Peer, id Id) {
	sort.Slice(peers, func(i, j int) bool {
		di := xor(id, peers[i].Id)
		dj := xor(id, peers[j].Id)
		return lt(di, dj)
	})
}