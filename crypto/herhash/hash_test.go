package herhash_test

import (
	"crypto/sha256"
	"testing"

	"github.com/herdius/herdius-core/crypto/herhash"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	testVector := []byte("herdius")
	hasher := herhash.New()
	hasher.Write(testVector)
	bz := hasher.Sum(nil)

	bz2 := herhash.Sum(testVector)

	hasher = sha256.New()
	hasher.Write(testVector)
	bz3 := hasher.Sum(nil)

	assert.Equal(t, bz, bz2)
	assert.Equal(t, bz, bz3)
}

func TestHashTruncated(t *testing.T) {
	testVector := []byte("herdius")
	hasher := herhash.NewTruncated()
	hasher.Write(testVector)
	bz := hasher.Sum(nil)

	bz2 := herhash.SumTruncated(testVector)

	hasher = sha256.New()
	hasher.Write(testVector)
	bz3 := hasher.Sum(nil)
	bz3 = bz3[:herhash.TruncatedSize]

	assert.Equal(t, bz, bz2)
	assert.Equal(t, bz, bz3)
}
