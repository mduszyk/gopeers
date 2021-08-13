package peers

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestAddFind(t *testing.T) {
	nodeId := big.NewInt(0)
	k := 20
	b := 5
	node := NewP2pNode(k, b, nodeId)
	for i := 0; i < k; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		peer := &Peer{id, nil, time.Now()}
		if !node.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
	}
	for i := 0; i < k; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		n := node.buckets.find(id)
		if !n.bucket.contains(id) {
			t.Errorf("bucket should contain given Id\n")
		}
	}
}

func TestBucketListSplit(t *testing.T) {
	nodeId := big.NewInt(0)
	k := 20
	b := 5
	node := NewP2pNode(k, b, nodeId)
	for i := 0; i < k+1; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		peer := &Peer{id, nil, time.Now()}
		if !node.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
	}
	if node.buckets.root.left == nil || node.buckets.root.right == nil {
		t.Errorf("there should be 2 buckets\n")
	}
	for i := 0; i < k+1; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		n := node.buckets.find(id)
		if !n.bucket.contains(id) {
			t.Errorf("bucket should contain given Id\n")
		}
	}
}
