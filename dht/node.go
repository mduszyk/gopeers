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

	result, err := peer.Proto.FindNode(node.Peer, node.Peer.Id)
	if err != nil {
		return err
	}
	for _, p := range result.peers {
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
		result, err := peer.Proto.FindNode(node.Peer, id)
		if err != nil {
			return err
		}
		for _, p := range result.peers {
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

func (node *KadNode) Lookup(id Id, findValue bool) (*FindResult, error) {
	peers := node.Tree.closest(id, node.k)

	seen := make(map[string]bool)
	for _, peer := range peers {
		key := string(peer.Id.Bytes())
		seen[key] = true
	}

	queried := make([]*Peer, 0, node.k)

	type poolResult struct {
		peer *Peer
		findResult *FindResult
	}

	input, output := Pool(node.alpha, func(payload Payload) (Payload, error) {
		peer := payload.(*Peer)
		var findResult *FindResult
		var err error
		if findValue {
			findResult, err = peer.Proto.FindValue(node.Peer, id)
		} else {
			findResult, err = peer.Proto.FindNode(node.Peer, id)
		}
		return poolResult{peer: peer, findResult: findResult}, err
	})
	defer close(input)

	n := min(node.alpha, len(peers))
	for _, peer := range peers[:n] {
		input <- peer
	}
	peers = peers[n:]
	in := n
	out := 0

	for out < in {
		result := <-output
		out += 1
		peer := result.value.(poolResult).peer

		if result.err != nil {
			log.Printf("FindNode failed: %v\n", result.err)
		} else {
			findResult := result.value.(poolResult).findResult
			if findResult.value != nil {
				return findResult, nil
			} else {
				queried = append(queried, peer)
				for _, p := range findResult.peers {
					key := string(p.Id.Bytes())
					if _, ok := seen[key]; !ok {
						peers = insertSorted(peers, p, id)
						seen[key] = true
					}
				}
			}
		}

		pending := in - out
		missing := node.k - len(queried)
		if len(peers) > 0 && pending < missing {
			input <- peers[0]
			peers = peers[1:]
			in += 1
		}
	}

	if findValue {
		return nil, errors.New("not found")
	}

	for _, p := range queried {
		peers = insertSorted(peers, p, id)
	}
	peers = peers[:min(node.k, len(peers))]
	result := &FindResult{peers: peers, value: nil}
	return result, nil
}

// Storage interface

func (node *KadNode) Set(key []byte, value []byte) error {
	id := BytesId(key)
	findResult, err := node.Lookup(id, false)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(len(findResult.peers))
	parallelize(findResult.peers, func(peer *Peer) {
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
	id := BytesId(key)
	findResult, err := node.Lookup(id, true)
	if err != nil {
		return nil, err
	}
	return findResult.value, nil
}

// Protocol interface

func (node *KadNode) Ping(sender *Peer, randomId Id) (Id, error) {
	node.add(sender)
	return randomId, nil
}

func (node *KadNode) FindNode(sender *Peer, id Id) (*FindResult, error) {
	node.add(sender)
	node.Tree.mutex.RLock()
	peers := node.Tree.closest(id, node.Tree.k)
	node.Tree.mutex.RUnlock()
	return &FindResult{peers: peers, value: nil}, nil
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
	log.Printf("Store, peer: %d, key: %d\n", node.Peer.Id, key)
	return node.Storage.Set(key.Bytes(), value)
}
