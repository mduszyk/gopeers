package dht

import (
	"errors"
	"github.com/mduszyk/gopeers/store"
	"log"
	"sync"
	"time"
)

type KadNode struct {
	k, b, alpha    int
	Peer *Peer
	Tree *bucketTree
	Storage store.Storage
}

func NewKadNode(k, b, alpha int, id Id, storage store.Storage) *KadNode {
	node := &KadNode{k: k, b: b, alpha: alpha, Tree: NewBucketTree(k), Storage: storage}
	node.Peer = &Peer{id, node, time.Now()}
	return node
}

func NewRandomIdKadNode(k, b, alpha int, storage store.Storage) (*KadNode, error) {
	nodeId, err := CryptoRandId()
	if err != nil {
		return nil, err
	}
	node := NewKadNode(k, b, alpha, nodeId, storage)
	return node, nil
}

func (node *KadNode) callPing(peer *Peer) error {
	randomId, err := CryptoRandId()
	if err != nil {
		return err
	}
	echoId, err := peer.Proto.Ping(node.Peer, randomId)
	if err != nil {
		return err
	}

	if !eq(echoId, randomId) {
		return errors.New("ping random id not echoed")
	}
	return nil
}

func (node *KadNode) add(peer *Peer) bool {
	peer.touch()
	node.Tree.mutex.Lock()
	n := node.Tree.Find(peer.Id)
	if i := n.Bucket.find(peer.Id); i > -1 {
		n.Bucket.remove(i)
		added := n.Bucket.add(peer)
		node.Tree.mutex.Unlock()
		return added
	} else if n.Bucket.isFull() {
		if n.Bucket.inRange(node.Peer.Id) || n.Bucket.depth % node.b != 0 {
			node.Tree.split(n)
			node.Tree.mutex.Unlock()
			return node.add(peer)
		} else {
			if j, leastSeenPeer := n.Bucket.leastSeen(); j > -1 {
				node.Tree.mutex.Unlock()
				err := node.callPing(leastSeenPeer)
				node.Tree.mutex.Lock()
				if k := n.Bucket.find(leastSeenPeer.Id); k > -1 {
					n.Bucket.remove(k)
				}
				if err != nil {
					node.Tree.mutex.Unlock()
					return node.add(peer)
				} else {
					leastSeenPeer.touch()
					n.Bucket.add(leastSeenPeer)
				}
			}
			node.Tree.mutex.Unlock()
			return false
		}
	} else {
		added := n.Bucket.add(peer)
		node.Tree.mutex.Unlock()
		return added
	}
}

func (node *KadNode) Join(peer *Peer) error {
	node.add(peer)

	peers, err := peer.Proto.FindNode(node.Peer, node.Peer.Id)
	if err != nil {
		return err
	}
	for _, p := range peers {
		node.add(p)
	}

	node.Tree.mutex.RLock()
	buckets := node.Tree.buckets(peer.Id)
	node.Tree.mutex.RUnlock()
	if len(buckets) > 1 {
		// skip bucket containing our id
		err = node.refreshBuckets(buckets[1:])
		if err != nil {
			return err
		}
	}

	return nil
}

func (node *KadNode) refresh(b *bucket) error {
	id := MathRandIdRange(b.lo, b.hi)

	for _, peer := range b.peers {
		peers, err := peer.Proto.FindNode(node.Peer, id)
		if err != nil {
			return err
		}
		for _, p := range peers {
			node.add(p)
		}
	}

	return nil
}

func (node *KadNode) refreshBuckets(buckets []*bucket) error {
	for _, b := range buckets {
		err := node.refresh(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (node *KadNode) RefreshAll() error {
	node.Tree.mutex.RLock()
	buckets := node.Tree.buckets(node.Peer.Id)
	node.Tree.mutex.RUnlock()
	return node.refreshBuckets(buckets)
}

func (node *KadNode) Lookup(id Id) []*Peer {
	peers := node.Tree.closest(id, node.k)

	seen := make(map[string]bool)
	for _, peer := range peers {
		key := string(peer.Id.Bytes())
		seen[key] = true
	}

	queried := make([]*Peer, 0, node.k)

	type result struct {
		peer *Peer
		found []*Peer
	}

	input, output := Pool(node.alpha, func(payload Payload) (Payload, error) {
		peer := payload.(*Peer)
		found, err := peer.Proto.FindNode(node.Peer, id)
		return result{peer, found}, err
	})
	defer close(input)

	n := min(node.alpha, len(peers))
	for _, peer := range peers[:n] {
		input <- peer
	}
	peers = peers[n:]
	inN := n
	outN := 0

	for outN < inN {
		r := <- output
		outN += 1
		peer := r.value.(result).peer

		if r.err != nil {
			log.Printf("FindNode failed: %v\n", r.err)
		} else {
			found := r.value.(result).found
			queried = append(queried, peer)
			for _, p := range found {
				key := string(p.Id.Bytes())
				if _, ok := seen[key]; !ok {
					peers = append(peers, p)
					seen[key] = true
				}
			}
			sortByDistance(peers, id)
		}

		if len(peers) > 0 && len(queried) < node.k {
			input <- peers[0]
			inN += 1
			peers = peers[1:]
		}
	}

	peers = append(peers, queried...)
	sortByDistance(peers, id)

	return peers[:min(node.k, len(peers))]
}

// Storage interface

func (node *KadNode) Set(key []byte, value []byte) error {
	id := BytesId(key)
	peers := node.Lookup(id)
	var wg sync.WaitGroup
	wg.Add(len(peers))
	parallelize(peers, func(peer *Peer) {
		err := peer.Proto.Store(node.Peer, id, value)
		if err != nil {
			log.Printf("Store failed, peer: %v, error: %v\n", peer, err)
		}
		wg.Done()
	})
	wg.Wait()
	// TODO what if all store fail
	return nil
}

func (node *KadNode) Get(key []byte) ([]byte, error) {
	// TODO
	return nil, nil
}

// Protocol interface

func (node *KadNode) Ping(sender *Peer, randomId Id) (Id, error) {
	node.add(sender)
	return randomId, nil
}

func (node *KadNode) FindNode(sender *Peer, id Id) ([]*Peer, error) {
	node.add(sender)
	node.Tree.mutex.RLock()
	peers := node.Tree.closest(id, node.Tree.k)
	node.Tree.mutex.RUnlock()
	return peers, nil
}

func (node *KadNode) FindValue(sender *Peer, key Id) (*FindResult, error) {
	value, err := node.Storage.Get(key.Bytes())
	if err != nil {
		node.Tree.mutex.RLock()
		peers := node.Tree.closest(key, node.Tree.k)
		node.Tree.mutex.RUnlock()
		result := &FindResult{value: nil, peers: peers}
		return result, nil
	}
	result := &FindResult{value: value, peers: nil}
	return result, nil
}

func (node *KadNode) Store(sender *Peer, key Id, value []byte) error {
	return node.Storage.Set(key.Bytes(), value)
}
