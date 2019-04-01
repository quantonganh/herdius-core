package ed_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/herdius/herdius-core/crypto"
	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"

	cmn "github.com/herdius/herdius-core/libs/common"
)

func TestSignAndValidateEd25519(t *testing.T) {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	addr40 := append([]byte{40}, pubKey.Address()...)

	hash2561 := sha256.Sum256(addr40)
	hash2562 := sha256.Sum256(hash2561[:])
	checksum := hash2562[:4]

	rawAddr := append(addr40, checksum...)
	herdiusAddr := base58.Encode(rawAddr)

	assert.Equal(t, 34, len(herdiusAddr))

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sig[7] ^= byte(0x01)

	fmt.Printf("Address is 1: %X\n", pubKey.Address())
	fmt.Printf("Address is 2: %X\n", addr40)
	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestSignAndValidateEd25519CreatingKeyFromSecret(t *testing.T) {
	salt := cmn.CreateRandSalt(52)
	var memory uint32
	memory = 32 * 1024
	// KDF : Argon2
	key := argon2.Key([]byte("Secret Passphrase"), []byte(salt), 3, memory, 4, 32)
	privKey := ed25519.GenPrivKeyFromSecret(key)

	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sig[7] ^= byte(0x01)

	assert.False(t, pubKey.VerifyBytes(msg, sig))

}

func TestCreateAddress(t *testing.T) {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	HERAddress := pubKey.GetAddress()

	assert.Equal(t, 34, len(HERAddress))
}
