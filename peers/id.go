package peers

import (
	"crypto/rand"
	"crypto/sha1"
	"math/big"
)

type Id = *big.Int

var maxId = new(big.Int).Lsh(big.NewInt(1), 160)

func RandomId() (Id, error) {
	id, err := rand.Int(rand.Reader, maxId)
	return id, err
}

func Sha1Id(data []byte) Id {
	hash := sha1.Sum(data)
	return new(big.Int).SetBytes(hash[:])
}