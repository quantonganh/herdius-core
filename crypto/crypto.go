package crypto

import (
	"github.com/herdius/herdius-core/crypto/herhash"
	cmn "github.com/herdius/herdius-core/libs/common"
)

const (
	// AddressSize is the size of a pubkey address.
	AddressSize = herhash.TruncatedSize
)

// Address : An address is a []byte, but hex-encoded even in JSON.
// []byte leaves us the option to change the address length.
// Use an alias so Unmarshal methods (with ptr receivers) are available too.
type Address = cmn.HexBytes

//AddressHash ...
func AddressHash(bz []byte) Address {
	return Address(herhash.SumTruncated(bz))
}

// PubKey ...
type PubKey interface {
	Address() Address
	Bytes() []byte
	VerifyBytes(msg []byte, sig []byte) bool
	Equals(PubKey) bool
	GetAddress() string // Creates address in string format and precedes the address with 'H'

}

// PrivKey ...
type PrivKey interface {
	Bytes() []byte
	Sign(msg []byte) ([]byte, error)
	PubKey() PubKey
	Equals(PrivKey) bool
}

// Symmetric ...
type Symmetric interface {
	Keygen() []byte
	Encrypt(plaintext []byte, secret []byte) (ciphertext []byte)
	Decrypt(ciphertext []byte, secret []byte) (plaintext []byte, err error)
}
