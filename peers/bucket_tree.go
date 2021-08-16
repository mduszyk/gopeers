package peers

import (
	"math/big"
	"sync"
)

type node struct {
	parent, left, right *node
	bucket *bucket
}

type bucketTree struct {
	k, size int
	root *node
	mutex *sync.RWMutex
}

func NewBucketTree(k int) *bucketTree {
	b := NewBucket(k, 0, big.NewInt(0), maxId)
	root := &node{bucket: b}
	return &bucketTree{k: k, size: 1, root: root, mutex: &sync.RWMutex{}}
}

func (tree *bucketTree) find(id Id) *node {
	n := tree.root
	ForeachBit(id, func(bit bool) bool {
		if bit {
			if n.right == nil {
				return false
			}
			n = n.right
		} else {
			if n.left == nil {
				return false
			}
			n = n.left
		}
		return true
	})
	return n
}

func (tree *bucketTree) split(n *node) {
	left, right := n.bucket.split()
	n.bucket = nil
	n.left = &node{parent: n, bucket: left}
	n.right = &node{parent: n, bucket: right}
	tree.size += 1
}

func (tree *bucketTree) closest(id Id, n int) []*Peer {
	node := tree.find(id)
	peers := make([]*Peer, n)
	closest := node.bucket.closest(id, n)
	copy(peers, closest)
	m := len(closest)
	child := node
	node = node.parent
	for node != nil && m < n {
		if child == node.left {
			m += appendRightPeers(peers[m:], node.right, id, n - m)
		} else {
			m += appendLeftPeers(peers[m:], node.left, id, n - m)
		}
		child = node
		node = node.parent
	}
	return peers[:m]
}

func appendRightPeers(peers []*Peer, node *node, id Id, n int) int {
	if node.bucket != nil {
		closest := node.bucket.closest(id, n)
		copy(peers, closest)
		return len(closest)
	} else {
		m := appendRightPeers(peers, node.left, id, n)
		m += appendRightPeers(peers, node.right, id, n - m)
		return m
	}
}

func appendLeftPeers(peers []*Peer, node *node, id Id, n int) int {
	if node.bucket != nil {
		closest := node.bucket.closest(id, n)
		copy(peers, closest)
		return len(closest)
	} else {
		m := appendLeftPeers(peers, node.right, id, n)
		m += appendLeftPeers(peers[m:], node.left, id, n - m)
		return m
	}
}

func (tree *bucketTree) buckets(id Id) []*bucket {
	node := tree.find(id)
	buckets := make([]*bucket, tree.size)
	buckets[0] = node.bucket
	child := node
	node = node.parent
	m := 1
	for node != nil {
		if child == node.left {
			m += appendRightBuckets(buckets[m:], node.right)
		} else {
			m += appendLeftBuckets(buckets[m:], node.left)
		}
		child = node
		node = node.parent
	}
	return buckets[:m]
}

func appendRightBuckets(buckets []*bucket, node *node) int {
	if node.bucket != nil {
		buckets[0] = node.bucket
		return 1
	} else {
		m := appendRightBuckets(buckets, node.left)
		m += appendRightBuckets(buckets[m:], node.right)
		return m
	}
}

func appendLeftBuckets(buckets []*bucket, node *node) int {
	if node.bucket != nil {
		buckets[0] = node.bucket
		return 1
	} else {
		m := appendLeftBuckets(buckets, node.right)
		m += appendLeftBuckets(buckets[m:], node.left)
		return m
	}
}
