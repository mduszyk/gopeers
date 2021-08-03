package peers

import (
	"math/big"
	"testing"
)

func TestInRange(t *testing.T) {
	lo := big.NewInt(0)
	hi := maxId
	bucket := NewBucket(20, lo, hi)
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
}
