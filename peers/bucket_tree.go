package peers

import (
	"math/big"
)

type node struct {
	parent *node
	left *node
	right *node
	bucket *bucket
}

func (n *node) split() {
	left, right := n.bucket.split()
	n.bucket = nil
	n.left = &node{parent: n, bucket: left}
	n.right = &node{parent: n, bucket: right}
}

type bucketTree struct {
	k int
	root *node
}

func NewBucketTree(k int) *bucketTree {
	b := NewBucket(k, 0, big.NewInt(0), maxId)
	root := &node{nil, nil, nil, b}
	return &bucketTree{k, root}
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

func appendRight(peers []*Peer, node *node, id Id, n int) int {
	if node.bucket != nil {
		closest := node.bucket.closest(id, n)
		copy(peers, closest)
		return len(closest)
	} else {
		m := appendRight(peers, node.left, id, n)
		m += appendRight(peers, node.right, id, n - m)
		return m
	}
}

func appendLeft(peers []*Peer, node *node, id Id, n int) int {
	if node.bucket != nil {
		closest := node.bucket.closest(id, n)
		copy(peers, closest)
		return len(closest)
	} else {
		m := appendLeft(peers, node.right, id, n)
		m += appendLeft(peers[m:], node.left, id, n - m)
		return m
	}
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
			m += appendRight(peers[m:], node.right, id, n - m)
		} else {
			m += appendLeft(peers[m:], node.left, id, n - m)
		}
		child = node
		node = node.parent
	}
	return peers[:m]
}
