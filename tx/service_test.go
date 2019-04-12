package tx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	"github.com/herdius/herdius-core/crypto/herhash"
	"github.com/herdius/herdius-core/crypto/merkle"
	cmn "github.com/herdius/herdius-core/libs/common"
	. "github.com/herdius/herdius-core/libs/test"
	"github.com/herdius/herdius-core/supervisor/transaction"
)

var service Service

func TestHashOfTx(t *testing.T) {
	service = GetTxsService()
	tx := getTx(1)
	txbz, err := cdc.MarshalJSON(tx)
	require.NoError(t, err, "Marshalling failed: %v.", err)
	assert.Equal(t, herhash.Sum(txbz), service.Hash(txbz))
}

func TestSimpleProofOfTxs(t *testing.T) {
	total := 100

	txs := make([][]byte, total)
	service = GetTxsService()
	for i := 0; i < total; i++ {
		tx := getTx(i + 1)
		txbz, err := cdc.MarshalJSON(tx)
		require.NoError(t, err, "Marshalling failed: %v.", err)
		assert.Equal(t, herhash.Sum(txbz), service.Hash(txbz))
		txs[i] = txbz
	}

	rootHash := merkle.SimpleHashFromByteSlices(txs)

	rootHash2, proofs := merkle.SimpleProofsFromByteSlices(txs)

	require.Equal(t, rootHash, rootHash2, "Unmatched root hashes: %X vs %X", rootHash, rootHash2)
	assert.NotNil(t, proofs)

	// For each tx, check the trail.
	for i, tx := range txs {
		txHash := herhash.Sum(tx)
		proof := proofs[i]

		// Check total/index
		require.Equal(t, proof.Index, i, "Unmatched indicies: %d vs %d", proof.Index, i)

		require.Equal(t, proof.Total, total, "Unmatched totals: %d vs %d", proof.Total, total)

		// Verify success
		err := proof.Verify(rootHash, txHash)
		require.NoError(t, err, "Verificatior failed: %v.", err)

		// Trail too long should make it fail
		origAunts := proof.Aunts
		proof.Aunts = append(proof.Aunts, cmn.RandBytes(32))
		err = proof.Verify(rootHash, txHash)
		require.Error(t, err, "Expected verification to fail for wrong trail length")

		proof.Aunts = origAunts

		// Trail too short should make it fail
		proof.Aunts = proof.Aunts[0 : len(proof.Aunts)-1]
		err = proof.Verify(rootHash, txHash)
		require.Error(t, err, "Expected verification to fail for wrong trail length")

		proof.Aunts = origAunts

		// Mutating the txHash should make it fail.
		err = proof.Verify(rootHash, MutateByteSlice(txHash))
		require.Error(t, err, "Expected verification to fail for mutated leaf hash")

		// Mutating the rootHash should make it fail.
		err = proof.Verify(MutateByteSlice(rootHash), txHash)
		require.Error(t, err, "Expected verification to fail for mutated root hash")

	}
}

func TestMerkleProofValidationOfTxs(t *testing.T) {
	total := 500

	txs := make([][]byte, total)
	service = GetTxsService()
	for i := 0; i < total; i++ {
		tx := getTx(i + 1)
		txbz, err := cdc.MarshalJSON(tx)
		require.NoError(t, err, "Marshalling failed: %v.", err)
		assert.Equal(t, herhash.Sum(txbz), service.Hash(txbz))
		txs[i] = txbz
	}
	service.SetTxs(txs)

	// Test Merkle Hash

	rootHash := service.MerkleHash()
	assert.NotNil(t, rootHash)

	// Verify Index returns the index of this transaction in the list
	for i := 0; i < total; i++ {
		actualIndex := service.Index(txs[i])
		assert.Equal(t, i, actualIndex)
	}

	// Verify IndexByHash where it should match the index of this transaction hash in the list
	for i := 0; i < total; i++ {
		actualIndex := service.IndexByHash(service.Hash(txs[i]))
		assert.Equal(t, i, actualIndex)
	}

	proofs := make([]Proof, total)

	for i := 0; i < total; i++ {
		proof := service.Proof(i)
		proofs[i] = proof
		var actualRootHash cmn.HexBytes
		actualRootHash = rootHash
		assert.Equal(t, proofs[i].RootHash, actualRootHash)
		err := proof.Validate(actualRootHash)

		require.NoError(t, err, "Validation failed: %v.", err)
	}
}

func getTx(nonce int) transaction.Tx {
	msg := []byte("Transfer 10 BTC")
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	sign, _ := privKey.Sign(msg)
	tx := transaction.Tx{
		Nonce:         uint64(nonce),
		Senderpubkey:  pubKey.Bytes(),
		Fee:           []byte("100"),
		Assetcategory: "Crypto",
		Assetname:     "BTC",
		Value:         []byte("10"),
		Signature:     sign,
	}

	return tx
}
