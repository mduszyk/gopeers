package peers

import (
	"crypto/sha1"
	"log"
	"math/big"
	"testing"
)

func TestSha1Id(t *testing.T) {
	data := []byte("Some test data.")
	hash := sha1.Sum(data)
	log.Printf("sha1: %x", hash)
	i := new(big.Int).SetBytes(hash[:])
	log.Printf("bigint: %d\n", i)
	id := Sha1Id(data)
	log.Printf("id: %d\n", id)
	if id.Cmp(i) != 0 {
		t.Errorf("failed creating sha1 id\n")
	}
}

func TestRandomId(t *testing.T) {
	id1, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random id, err: %v\n", err)
	}
	id2, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random id, err: %v\n", err)
	}
	log.Printf("id1: %d\n", id1)
	log.Printf("id2: %d\n", id2)
	if id1.Cmp(id2) == 0 {
		t.Errorf("generated two equal random ids\n")
	}
}
