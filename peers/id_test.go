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
	log.Printf("Id: %d\n", id)
	if id.Cmp(i) != 0 {
		t.Errorf("failed creating sha1 Id\n")
	}
}

func TestRandomId(t *testing.T) {
	id1, err := CryptoRandId()
	if err != nil {
		t.Errorf("failed generating random Id, err: %v\n", err)
	}
	id2, err := CryptoRandId()
	if err != nil {
		t.Errorf("failed generating random Id, err: %v\n", err)
	}
	log.Printf("id1: %d\n", id1)
	log.Printf("id2: %d\n", id2)
	if id1.Cmp(id2) == 0 {
		t.Errorf("generated two equal random ids\n")
	}
}

func intBits(trueBits []uint) *big.Int {
	i := big.NewInt(0)
	for _, bit := range trueBits {
		i.Add(i, new(big.Int).Lsh(big.NewInt(1), bit))
	}
	return i
}

func TestForeachBit(t *testing.T) {
	id := intBits([]uint{IdBits - 1, IdBits - 3})
	bits := make([]bool, 0, IdBits)
	ForeachBit(id, func(bit bool) bool {
		bits = append(bits, bit)
		return true
	})
	if len(bits) != IdBits {
		t.Errorf("invalid count of bits: %d\n", len(bits))
	}
	if !bits[0] || bits[1] || !bits[2] {
		t.Errorf("invalid bits iteration\n")
	}
	id = intBits([]uint{IdBits / 2})
	bits = make([]bool, 0, IdBits)
	ForeachBit(id, func(bit bool) bool {
		bits = append(bits, bit)
		return true
	})
	if len(bits) != IdBits {
		t.Errorf("invalid count of bits: %d\n", len(bits))
	}
	id = intBits([]uint{70})
	bits = make([]bool, 0, IdBits)
	ForeachBit(id, func(bit bool) bool {
		bits = append(bits, bit)
		return true
	})
	if len(bits) != IdBits {
		t.Errorf("invalid count of bits: %d\n", len(bits))
	}
	id = intBits([]uint{10})
	bits = make([]bool, 0, IdBits)
	ForeachBit(id, func(bit bool) bool {
		bits = append(bits, bit)
		return true
	})
	if len(bits) != IdBits {
		t.Errorf("invalid count of bits: %d\n", len(bits))
	}
	id = intBits([]uint{IdBits * 2, IdBits - 1, IdBits - 3})
	bits = make([]bool, 0, IdBits)
	ForeachBit(id, func(bit bool) bool {
		bits = append(bits, bit)
		return true
	})
	if len(bits) != IdBits {
		t.Errorf("invalid count of bits: %d\n", len(bits))
	}
	if !bits[0] || bits[1] || !bits[2] {
		t.Errorf("invalid bits iteration\n")
	}
	id = big.NewInt(0)
	n := 0
	ForeachBit(id, func(bit bool) bool {
		if bit {
			t.Errorf("all bits should be false\n")
		}
		n++
		return true
	})
	if n != IdBits {
		t.Errorf("foreach didn't go through all bits\n")
	}
}

func TestRandomIdRange(t *testing.T) {
	lo := big.NewInt(10)
	hi := big.NewInt(11)
	id, err := CryptoRandIdRange(lo, hi)
	if err != nil {
		t.Errorf("failed generating id: %v\n", err)
	}
	if !eq(id, lo) {
		t.Errorf("invalid id generated\n")
	}
}
