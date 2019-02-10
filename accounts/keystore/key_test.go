package keystore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/herdius/herdius-core/crypto"
	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestLoadOrGenNodeKey(t *testing.T) {
	filePath := filepath.Join(os.TempDir(), cmn.RandStr(12)+"_herdius_key.json")

	key, err := StoreKey(filePath)
	assert.Nil(t, err)

	key2, err := LoadKeyUsingPrivKey(filePath)
	assert.Nil(t, err)

	assert.Equal(t, key, key2)

	privKey := (*key).PrivKey
	pubKey := privKey.PubKey()

	privKey2 := (*key2).PrivKey
	pubKey2 := privKey2.PubKey()

	assert.Equal(t, pubKey, pubKey2)
	assert.Equal(t, pubKey.Address(), pubKey2.Address())
}

func TestLoadKeyFromJSONFile(t *testing.T) {

	salt := "LKtAMSAoJuqdfHpAExnfgdMuwvtvBejZtDgBCoHMgitcnSMnIfIh"
	expectedPubKeyAddress := "C8123FA11D5D0B4FB8E25F2FA90F88C64CF4670C"
	var memory uint32
	memory = 32 * 1024
	key := argon2.Key([]byte("Secret Passphrase"), []byte(salt), 3, memory, 4, 32)

	// Expected Private Key
	expPrivKey := ed25519.GenPrivKeyFromSecret(key)

	privKey, err := LoadKeyUsingSecretKeyAndSalt("./testdata/v1_test_argon2.json")
	assert.Nil(t, err)

	assert.True(t, expPrivKey.Equals((*privKey).PrivKey))

	assert.Equal(t, (*privKey).PrivKey.PubKey(), expPrivKey.PubKey())

	address := (*privKey).PrivKey.PubKey().Address().String()
	assert.Equal(t, expectedPubKeyAddress, address)

}

func TestSignAndValidateEd25519(t *testing.T) {
	key, err := LoadKeyUsingSecretKeyAndSalt("./testdata/v1_test_argon2.json")
	assert.Nil(t, err)
	assert.NotEmpty(t, key)

	privKey := (*key).PrivKey
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
