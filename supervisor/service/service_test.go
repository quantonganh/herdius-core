package service

import (
	"testing"

	ed25519 "github.com/herdius/herdius-core/crypto/ed"

	"github.com/herdius/herdius-core/crypto/secp256k1"
	"github.com/herdius/herdius-core/supervisor/transaction"
	"github.com/stretchr/testify/assert"
)

func TestCreateTxBatchesFromFile(t *testing.T) { /*
		total := 200
		supsvc := &Supervisor{}
		supsvc.SetWriteMutex()
		supsvc.CreateTxBatchesFromFile("../testdata/txs.json", 5, total)
		actualNoOfBatches := len(*(supsvc.TxBatches))
		expectedNoOfBatches := 5
		assert.Equal(t, expectedNoOfBatches, actualNoOfBatches, "Expected number of batches should be 5")

		batches := *supsvc.TxBatches

		counter := 0

		var txService transaction.Service

		for _, batch := range batches {
			txs := make([][]byte, 0)
			assert.NotNil(t, batch)
			assert.Equal(t, 200, len(batch), "Expected number of transactions should be 200")

			txService = transaction.TxService()
			for i := 0; i < 200; i++ {
				txbz := batch[i]

				tx := transaction.Tx{}
				cdc.UnmarshalJSON(txbz, &tx)

				txService.AddTx(tx)
				assert.Equal(t, uint64(counter+1), tx.Nonce, "Expected Transaction nonce did not match.")
				counter++

			}

			// Create Child block from the batch
			txList := *(txService.GetTxList())
			assert.NotNil(t, txList)
			assert.Equal(t, 200, len(txList.Transactions))

			supsvc := &Supervisor{}
			supsvc.SetWriteMutex()
			cb := supsvc.CreateChildBlock(nil, &txList, 1, []byte{0})

			assert.NotNil(t, cb)
			rootHash := cb.GetHeader().GetRootHash()
			assert.NotNil(t, rootHash)

			txs = cb.GetTxsData().Tx

			rootHash2, proofs := merkle.SimpleProofsFromByteSlices(txs)

			require.Equal(t, rootHash, rootHash2, "Unmatched root hashes: %X vs %X", rootHash, rootHash2)
			assert.NotNil(t, proofs)

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
				err = proof.Verify(rootHash, ctest.MutateByteSlice(txHash))
				require.Error(t, err, "Expected verification to fail for mutated leaf hash")

				// Mutating the rootHash should make it fail.
				err = proof.Verify(ctest.MutateByteSlice(rootHash), txHash)
				require.Error(t, err, "Expected verification to fail for mutated root hash")
			}

		} */
}

func TestRemoveValidator(t *testing.T) {
	supsvc := &Supervisor{}
	supsvc.SetWriteMutex()
	supsvc.AddValidator([]byte{1}, "add-01")
	supsvc.AddValidator([]byte{1}, "add-02")
	supsvc.AddValidator([]byte{1}, "add-03")
	supsvc.AddValidator([]byte{1}, "add-04")
	supsvc.AddValidator([]byte{1}, "add-05")
	supsvc.AddValidator([]byte{1}, "add-06")
	supsvc.AddValidator([]byte{1}, "add-07")
	supsvc.AddValidator([]byte{1}, "add-08")
	supsvc.AddValidator([]byte{1}, "add-09")
	supsvc.AddValidator([]byte{1}, "add-10")

	assert.Equal(t, 10, len(supsvc.Validator))

	supsvc.RemoveValidator("add-04")
	supsvc.RemoveValidator("add-08")
	supsvc.RemoveValidator("add-10")
	assert.Equal(t, 7, len(supsvc.Validator))

}

func TestCreateChildBlock(t *testing.T) {
	var txService transaction.Service
	txService = transaction.TxService()
	for i := 1; i <= 200; i++ {
		tx := getTx(i)
		txService.AddTx(tx)
	}
	txList := txService.GetTxList()
	assert.NotNil(t, txList)
	assert.Equal(t, 200, len((*txList).Transactions))

	supsvc := &Supervisor{}
	supsvc.SetWriteMutex()
	cb := supsvc.CreateChildBlock(nil, txList, 1, []byte{0})

	assert.NotNil(t, cb)
}

func getTx(nonce int) transaction.Tx {
	msg := []byte("Transfer 10 BTC")
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	sign, _ := privKey.Sign(msg)
	asset := transaction.Asset{
		Nonce:    string(nonce),
		Fee:      "100",
		Category: "Crypto",
		Symbol:   "BTC",
		Value:    "10",
		Network:  "Herdius",
	}
	tx := transaction.Tx{
		SenderPubKey:  string(pubKey.Bytes()),
		SenderAddress: string(pubKey.Address()),
		Asset:         asset,
		Signature:     string(sign),
		Type:          "update",
	}

	return tx
}

func TestCreateChildBlockForSecp256k1Account(t *testing.T) {
	var txService transaction.Service
	txService = transaction.TxService()
	for i := 1; i <= 200; i++ {
		tx := getTxSecp256k1Account(i)
		txService.AddTx(tx)
	}
	txList := txService.GetTxList()
	assert.NotNil(t, txList)
	assert.Equal(t, 200, len((*txList).Transactions))

	supsvc := &Supervisor{}
	supsvc.SetWriteMutex()
	cb := supsvc.CreateChildBlock(nil, txList, 1, []byte{0})

	assert.NotNil(t, cb)
}

func getTxSecp256k1Account(nonce int) transaction.Tx {
	msg := []byte("Transfer 10 BTC")
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	sign, _ := privKey.Sign(msg)
	asset := transaction.Asset{
		Nonce:    string(nonce),
		Fee:      "100",
		Category: "Crypto",
		Symbol:   "BTC",
		Value:    "10",
		Network:  "Herdius",
	}
	tx := transaction.Tx{
		SenderPubKey:  string(pubKey.Bytes()),
		SenderAddress: string(pubKey.Address()),
		Asset:         asset,
		Signature:     string(sign),
		Type:          "update",
	}
	// pp, _ := transaction.PrettyPrint(tx)
	// fmt.Println(pp)
	return tx
}
