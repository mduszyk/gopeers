package peers

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestInRange(t *testing.T) {
	lo := big.NewInt(0)
	hi := maxId
	bucket := NewBucket(20, 0, lo, hi)
	if !bucket.inRange(lo) {
		t.Errorf("low should be in range\n")
	}
	if bucket.inRange(hi) {
		t.Errorf("hi should not be in range\n")
	}
	a := new(big.Int).Sub(lo, big.NewInt(1))
	if bucket.inRange(a) {
		t.Errorf("value %d should not be in range\n", a)
	}
	b := new(big.Int).Sub(hi, big.NewInt(1))
	if !bucket.inRange(b) {
		t.Errorf("value %d should be in range\n", b)
	}
	c := new(big.Int).Add(hi, big.NewInt(1))
	if bucket.inRange(c) {
		t.Errorf("value %d should not be in range\n", c)
	}
	d, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random Id: %v\n", err)
	}
	if !bucket.inRange(d) {
		t.Errorf("random Id %d should be in range\n", d)
	}
	e := Sha1Id([]byte("test0"))
	if !bucket.inRange(e) {
		t.Errorf("sha1 Id %d should be in range\n", e)
	}

}

func TestAdd(t *testing.T) {
	lo := big.NewInt(0)
	hi := maxId
	bucket := NewBucket(20, 0, lo, hi)
	id := Sha1Id([]byte("test123"))
	peer := &Peer{id, nil, time.Now()}
	if !bucket.add(peer) {
		t.Errorf("bucket should add peer\n")
	}
	if !bucket.contains(id) {
		t.Errorf("bucket should contain added peer\n")
	}
}

func TestFull(t *testing.T) {
	k := 20
	lo := big.NewInt(0)
	hi := maxId
	bucket := NewBucket(k, 0, lo, hi)
	for i := 0; i < k; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		peer := &Peer{id, nil, time.Now()}
		if !bucket.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
	}
	if !bucket.isFull() {
		t.Errorf("bucket should be full\n")
	}
	id := Sha1Id([]byte("test123"))
	peer := &Peer{id, nil, time.Now()}
	if bucket.add(peer) {
		t.Errorf("bucket should not add peer\n")
	}
}

func TestRemove(t *testing.T) {
	lo := big.NewInt(0)
	hi := maxId
	bucket := NewBucket(20, 0, lo, hi)
	for i := 0; i < 10; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		peer := &Peer{id, nil, time.Now()}
		if !bucket.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
	}
	id := Sha1Id([]byte("test5"))
	if !bucket.contains(id) {
		t.Errorf("bucket should contain peer\n")
	}
	if !bucket.remove(5) {
		t.Errorf("bucket should remove peer\n")
	}
	if bucket.contains(id) {
		t.Errorf("bucket should not contain peer\n")
	}
}

func TestFind(t *testing.T) {
	lo := big.NewInt(0)
	hi := maxId
	bucket := NewBucket(20, 0, lo, hi)
	for i := 0; i < 10; i++ {
		id := Sha1Id([]byte(fmt.Sprintf("test%d", i)))
		peer := &Peer{id, nil, time.Now()}
		if !bucket.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
	}
	id := Sha1Id([]byte("test5"))
	i := bucket.find(id)
	if i != 5 {
		t.Errorf("bucket should find peer, position: %d\n", i)
	}
}

//func TestDepth(t *testing.T) {
//	id1 := intBits([]uint{191, 190, 188, 186, 74, 1})
//	id2 := intBits([]uint{191, 190, 188, 180, 74, 1})
//	id3 := intBits([]uint{191, 190, 188, 161, 74, 1})
//	lo := big.NewInt(0)
//	hi := maxId
//	bucket := NewBucket(20, lo, hi)
//	bucket.add(&Peer{id1, nil, time.Now()})
//	bucket.add(&Peer{id2, nil, time.Now()})
//	bucket.add(&Peer{id3, nil, time.Now()})
//	depth := bucket.depth()
//	expected := 5
//	if depth != expected {
//		t.Errorf("incorrect depth: %d, expected: %d\n", depth, expected)
//	}
//}

func TestSplit(t *testing.T) {
	k := 20
	bucket := NewBucket(k, 0, big.NewInt(0), maxId)
	base := big.NewInt(0)
	for i := 0; i < k; i++ {
		id := new(big.Int).Add(base, big.NewInt(int64(i)))
		peer := &Peer{id, nil, time.Now()}
		if !bucket.add(peer) {
			t.Errorf("bucket should add peer %d\n", i)
		}
		if i == 9 {
			base = intBits([]uint{159})
		}
	}
	bucket1, bucket2 := bucket.split()
	if len(bucket1.peers) != k/2 {
		t.Errorf("bucket should contain half of the elements\n")
	}
	base = big.NewInt(0)
	for i := 0; i < k/2; i++ {
		id := new(big.Int).Add(base, big.NewInt(int64(i)))
		if !bucket1.contains(id) {
			t.Errorf("bucket should contain Id: %d\n", id)
		}
	}
	if len(bucket2.peers) != k/2 {
		t.Errorf("bucket should contain half of the elements\n")
	}
	base = intBits([]uint{159})
	for i := 10; i < k; i++ {
		id := new(big.Int).Add(base, big.NewInt(int64(i)))
		if !bucket2.contains(id) {
			t.Errorf("bucket should contain Id: %d\n", id)
		}
	}
}
