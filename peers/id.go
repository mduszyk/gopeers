package peers

import (
	"crypto/rand"
	"crypto/sha1"
	"math/big"
	"math/bits"
)

type Id = *big.Int

const IdBits = 160

var maxId = new(big.Int).Lsh(big.NewInt(1), IdBits)

func RandomId() (Id, error) {
	return rand.Int(rand.Reader, maxId)
}

func Sha1Id(data []byte) Id {
	hash := sha1.Sum(data)
	return new(big.Int).SetBytes(hash[:])
}

func ForeachBit(id Id, f func(bit bool) bool) {
	words := id.Bits()
	skipBits := len(words) * bits.UintSize - IdBits
	skipWords := 0
	if skipBits > 0 {
		skipWords = skipBits / bits.UintSize
		skipBits = skipBits - skipWords * bits.UintSize
	}
	for i := len(words) - 1 - skipWords; i >= 0; i-- {
		word := words[i]
		for j := 1 + skipBits; j <= bits.UintSize; j++ {
			mask := big.Word(1) << (bits.UintSize - j)
			if !f(word & mask > 0) {
				return
			}
		}
		skipBits = 0
	}
}

func xor(a Id, b Id) Id {
	return new(big.Int).Xor(a, b)
}

func eq(a Id, b Id) bool {
	return a.Cmp(b) == 0
}

func lt(a Id, b Id) bool {
	return a.Cmp(b) == -1
}

func lte(a Id, b Id) bool {
	r := a.Cmp(b)
	return r == 0 || r == -1
}
