package crypto

import (
	crand "crypto/rand"
	"io"
)

// This only uses the OS's randomness
func randBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := crand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

// This only uses the OS's randomness
func CRandBytes(numBytes int) []byte {
	return randBytes(numBytes)
}

// CReader : Returns a crand.Reader.
func CReader() io.Reader {
	return crand.Reader
}
