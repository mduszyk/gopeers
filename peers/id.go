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
	for i := len(words) - 1; i >= 0; i-- {
		word := words[i]
		for j := 1; j <= bits.UintSize; j++ {
			mask := big.Word(1) << (bits.UintSize - j)
			bools = append(bools, word & mask > 0)
		}
	}
	return bools
}

func SharedBits(bools []bool, id Id) []bool {
	if len(bools) == 0 {
		return bools
	}
	n := 0
	words := id.Bits()
	for i := len(words) - 1; i >= 0; i-- {
		word := words[i]
		for j := 1; j <= bits.UintSize; j++ {
			mask := big.Word(1) << (bits.UintSize - j)
			bit := word & mask > 0
			if n < len(bools) && bit == bools[n] {
				n++
			} else {
				return bools[:n]
			}
		}
	}
	return bools[:n]
}
