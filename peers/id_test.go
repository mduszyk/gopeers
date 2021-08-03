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

func contains(s []uint, e uint) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func TestToBits(t *testing.T) {
	id := big.NewInt(0)
	powers := []uint{2, 3, 7, 16, 65, 128, 160}
	for _, p := range powers {
		id.Add(id, new(big.Int).Lsh(big.NewInt(1), p))
	}
	bits := ToBits(id)
	for i, bit := range bits {
		bitIndex := len(bits) - 1 - i
		if contains(powers, uint(bitIndex)) {
			if !bit {
				t.Errorf("bit %d should be set\n", bitIndex)
			}
		} else {
			if bit {
				t.Errorf("bit %d should not be set\n", bitIndex)
			}
		}
	}
}