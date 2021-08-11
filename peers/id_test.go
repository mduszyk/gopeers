package peers

import (
	"crypto/sha1"
	"log"
	"math/big"
	"reflect"
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
	id1, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random Id, err: %v\n", err)
	}
	id2, err := RandomId()
	if err != nil {
		t.Errorf("failed generating random Id, err: %v\n", err)
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

func intBits(trueBits []uint) *big.Int {
	i := big.NewInt(0)
	for _, bit := range trueBits {
		i.Add(i, new(big.Int).Lsh(big.NewInt(1), bit))
	}
	return i
}

func TestToBits(t *testing.T) {
	trueBits := []uint{2, 3, 7, 16, 65, 128, 160}
	id := intBits(trueBits)
	bits := ToBits(id)
	for i, bit := range bits {
		bitIndex := len(bits) - 1 - i
		if contains(trueBits, uint(bitIndex)) {
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

func TestSharedBits(t *testing.T) {
	id1 := intBits([]uint{159, 158, 156, 154, 74, 1})
	id2 := intBits([]uint{159, 158, 156, 154, 74, 1})
	id3 := intBits([]uint{159, 158, 156, 153, 74, 1})
	prefix := ToBits(id1)
	prefix = SharedBits(prefix, id2)
	prefix = SharedBits(prefix, id3)
	expected := []bool{true, true, false, true, false}
	if !reflect.DeepEqual(prefix, expected) {
		t.Errorf("incorrect prefix %v, expected %v\n", prefix, expected)
	}
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
}