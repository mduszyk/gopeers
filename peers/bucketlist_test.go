package peers

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestAddFind(t *testing.T) {
	nodeId := big.NewInt(0)
	nodePeer := Peer{nodeId, nil, time.Now()}
	k := 20
	splitLevelB := 5
	bucketList := NewBucketList(k, splitLevelB, nodePeer)
	for i := 0; i < k; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		peer := Peer{id, nil, time.Now()}
		if !bucketList.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
	}
	for i := 0; i < k; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		j, b := bucketList.find(id)
		if j < 0 {
			t.Errorf("bucket list should find bucket containing given id\n")
		}
		if !b.contains(id) {
			t.Errorf("bucket should contain given id\n")
		}
	}
}
