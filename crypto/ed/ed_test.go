package ed_test

import (
	"testing"

	"github.com/herdius/herdius-core/crypto"
	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"

	cmn "github.com/herdius/herdius-core/libs/common"
)

func TestSignAndValidateEd25519(t *testing.T) {
	privKey := ed25519.GenPrivKey()
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
