package transaction_test

import (
	"testing"

	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	"github.com/herdius/herdius-core/supervisor/transaction"
	"github.com/stretchr/testify/assert"
)

func TestAddTx(t *testing.T) {

	tx := getTx(1)

	var txService transaction.Service
	txService = transaction.TxService()
	txService.AddTx(tx)

	txList := txService.GetTxList()
	assert.NotNil(t, txList)
	assert.Equal(t, 1, len((*txList).Transactions))
	assert.Equal(t, uint64(1), (*txList).Transactions[0].Nonce)
	assert.Equal(t, tx.Senderpubkey, (*txList).Transactions[0].Senderpubkey)
	assert.Equal(t, tx.Signature, (*txList).Transactions[0].Signature)
}

func TestGetTxList(t *testing.T) {
	var txService transaction.Service
	txService = transaction.TxService()
	for i := 1; i <= 10; i++ {
		tx := getTx(i)
		txService.AddTx(tx)
	}
	txList := txService.GetTxList()
	assert.NotNil(t, txList)
	assert.Equal(t, 10, len((*txList).Transactions))
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
