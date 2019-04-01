package cryptoAmino

import (
	"fmt"
	"os"
	"testing"

	"github.com/herdius/herdius-core/crypto"

	"github.com/herdius/herdius-core/crypto/secp256k1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type byter interface {
	Bytes() []byte
}

func checkAminoBinary(t *testing.T, src, dst interface{}, size int) {
	bz, err := cdc.MarshalBinaryBare(src)

	require.Nil(t, err, "%+v", err)
	if byterSrc, ok := src.(byter); ok {
		// Make sure this is compatible with current (Bytes()) encoding.
		assert.Equal(t, byterSrc.Bytes(), bz, "Amino binary vs Bytes() mismatch")
	}

	// Make sure we have the expected length.
	if size != -1 {
		assert.Equal(t, size, len(bz), "Amino binary size mismatch")
	}
	// Unmarshal.
	err = cdc.UnmarshalBinaryBare(bz, dst)
	require.Nil(t, err, "%+v", err)
}

func checkAminoJSON(t *testing.T, src interface{}, dst interface{}, isNil bool) {
	// Marshal to JSON bytes.
	js, err := cdc.MarshalJSON(src)
	require.Nil(t, err, "%+v", err)
	if isNil {
		assert.Equal(t, string(js), `null`)
	} else {
		assert.Contains(t, string(js), `"type":`)
		assert.Contains(t, string(js), `"value":`)
	}
	// Unmarshal.
	err = cdc.UnmarshalJSON(js, dst)
	require.Nil(t, err, "%+v", err)

	fmt.Printf("dst\n%v\n", dst)
}

func ExamplePrintRegisteredTypes() {
	cdc.PrintTypes(os.Stdout)
	// Output: | Type | Name | Prefix | Length | Notes |
	//| ---- | ---- | ------ | ----- | ------ |
	//| PubKeySecp256k1 | herdius/PubKeySecp256k1 | 0xF181CAB8 | 0x21 |  |
	//| PrivKeySecp256k1 | herdius/PrivKeySecp256k1 | 0x953F313C | 0x20 |  |
}

func TestKeyEncodings(t *testing.T) {
	cases := []struct {
		privKey                    crypto.PrivKey
		privSize, pubSize, sigSize int // binary sizes
	}{
		{
			privKey:  secp256k1.GenPrivKey(),
			privSize: 37,
			pubSize:  38,
			sigSize:  65,
		},
	}

	for tcIndex, tc := range cases {
		// Check (de/en)codings of PrivKeys.
		var priv2, priv3 crypto.PrivKey
		checkAminoBinary(t, tc.privKey, &priv2, tc.privSize)
		assert.EqualValues(t, tc.privKey, priv2, "tc #%d", tcIndex)

		checkAminoJSON(t, tc.privKey, &priv3, false)
		assert.EqualValues(t, tc.privKey, priv3, "tc #%d", tcIndex)

		// Check (de/en)codings of Signatures.
		var sig1, sig2 []byte
		sig1, err := tc.privKey.Sign([]byte("something"))
		assert.NoError(t, err, "tc #%d", tcIndex)
		checkAminoBinary(t, sig1, &sig2, tc.sigSize)
		assert.EqualValues(t, sig1, sig2, "tc #%d", tcIndex)

		// Check (de/en)codings of PubKeys.
		pubKey := tc.privKey.PubKey()

		var pub2 crypto.PubKey
		var pub3 crypto.PubKey
		checkAminoBinary(t, pubKey, &pub2, tc.pubSize)
		assert.EqualValues(t, pubKey, pub2, "tc #%d", tcIndex)
		checkAminoJSON(t, pubKey, &pub3, false)
		assert.EqualValues(t, pubKey, pub3, "tc #%d", tcIndex)
	}
}

func TestNilEncodings(t *testing.T) {

	// Check nil Signature.
	var a, b []byte
	checkAminoJSON(t, &a, &b, true)
	assert.EqualValues(t, a, b)

	// Check nil PubKey.
	var c, d crypto.PubKey
	checkAminoJSON(t, &c, &d, true)
	assert.EqualValues(t, c, d)

	// Check nil PrivKey.
	var e, f crypto.PrivKey
	checkAminoJSON(t, &e, &f, true)
	assert.EqualValues(t, e, f)
}

func TestPubKeyInvalidDataProperReturnsEmpty(t *testing.T) {
	pk, err := PubKeyFromBytes([]byte("foo"))
	require.NotNil(t, err)
	require.Nil(t, pk)
}

func TestPubkeyAminoRoute(t *testing.T) {
	tests := []struct {
		key   crypto.PubKey
		want  string
		found bool
	}{
		{secp256k1.PubKeySecp256k1{}, secp256k1.PubKeyAminoName, true},
	}
	for i, tc := range tests {
		got, found := PubkeyAminoRoute(cdc, tc.key)
		require.Equal(t, tc.found, found, "not equal on tc %d", i)
		if tc.found {
			require.Equal(t, tc.want, got, "not equal on tc %d", i)
		}
	}

}
