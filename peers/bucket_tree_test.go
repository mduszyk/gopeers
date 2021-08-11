package peers

import (
	"testing"
)

func TestTreeFind(t *testing.T) {
	tree := NewBucketTree(20)
	tree.root.split()
	tree.root.left.split()
	tree.root.right.split()
	peer := &Peer{Id: intBits([]uint{IdBits - 1, IdBits - 2})}
	n := tree.find(peer.Id)
	if n != tree.root.right.right {
		t.Errorf("found wrong node for id: %0160b\n", peer.Id)
	}
	peer = &Peer{Id: intBits([]uint{IdBits - 1})}
	n = tree.find(peer.Id)
	if n != tree.root.right.left {
		t.Errorf("found wrong node for id: %0160b\n", peer.Id)
	}
	peer = &Peer{Id: intBits([]uint{IdBits - 2})}
	n = tree.find(peer.Id)
	if n != tree.root.left.right {
		t.Errorf("found wrong node for id: %0160b\n", peer.Id)
	}
	peer = &Peer{Id: intBits([]uint{0})}
	n = tree.find(peer.Id)
	if n != tree.root.left.left {
		t.Errorf("found wrong node for id: %0160b\n", peer.Id)
	}
}
