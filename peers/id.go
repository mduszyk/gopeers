package peers

import (
	"crypto/rand"
	"crypto/sha1"
	"math/big"
	"math/bits"
)

type Id = *big.Int

var maxId = new(big.Int).Lsh(big.NewInt(1), 160)

func RandomId() (Id, error) {
	return rand.Int(rand.Reader, maxId)
}

func Sha1Id(data []byte) Id {
	hash := sha1.Sum(data)
	return new(big.Int).SetBytes(hash[:])
}

func ToBits(id Id) []bool {
	words := id.Bits()
	bools := make([]bool, 0, bits.UintSize * len(words))
	ForeachBit(id, func(bit bool) bool {
		bools = append(bools, bit)
		return true
	})
	return bools
}

func SharedBits(bools []bool, id Id) []bool {
	if len(bools) == 0 {
		return bools
	}
	n := 0
	ForeachBit(id, func(bit bool) bool {
		if n < len(bools) && bit == bools[n] {
			n++
		} else {
			return false
		}
		return true
	})
	return bools[:n]
}

func ForeachBit(id Id, f func(bit bool) bool) {
	words := id.Bits()
	for i := len(words) - 1; i >= 0; i-- {
		word := words[i]
		for j := 1; j <= bits.UintSize; j++ {
			mask := big.Word(1) << (bits.UintSize - j)
			bit := word & mask > 0
			if !f(bit) {
				return
			}
		}
	}
}
