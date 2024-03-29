package dht

import (
	"crypto/rand"
	"crypto/sha1"
	"math/big"
	"math/bits"
	mathRand "math/rand"
	"time"
)

type Id = *big.Int

const IdBits = 160

var maxId = new(big.Int).Lsh(big.NewInt(1), IdBits)

func MathRandId() Id {
	rnd := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	return new(big.Int).Rand(rnd, maxId)
}

func MathRandIdRange(lo Id, hi Id) Id {
	d := new(big.Int).Sub(hi, lo)
	r := mathRand.New(mathRand.NewSource(1))
	id := new(big.Int).Rand(r, d)
	return new(big.Int).Add(lo, id)
}

func CryptoRandId() (Id, error) {
	return rand.Int(rand.Reader, maxId)
}

func CryptoRandIdRange(lo Id, hi Id) (Id, error) {
	d := new(big.Int).Sub(hi, lo)
	id, err := rand.Int(rand.Reader, d)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Add(lo, id), nil
}

func Sha1Id(data []byte) Id {
	hash := sha1.Sum(data)
	return new(big.Int).SetBytes(hash[:])
}

func BytesId(bytes []byte) Id {
	return new(big.Int).SetBytes(bytes)
}

func ForeachBit(id Id, f func(bit bool) bool) {
	words := id.Bits()
	// zero has special representation
	if words == nil {
		for i := 0; i < IdBits; i++ {
			if !f(false) {
				return
			}
		}
		return
	}

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
