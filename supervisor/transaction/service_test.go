package transaction_test

import (
	"testing"

	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	"github.com/herdius/herdius-core/crypto/secp256k1"
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
	assert.Equal(t, string(1), (*txList).Transactions[0].Asset.Nonce)
	assert.Equal(t, tx.SenderPubKey, (*txList).Transactions[0].SenderPubKey)
	assert.Equal(t, tx.Signature, (*txList).Transactions[0].Signature)
}

func TestAddTxSecp256k1(t *testing.T) {

	tx := getTxUsingSecp256k1Account(1)

	var txService transaction.Service
	txService = transaction.TxService()
	txService.AddTx(tx)

	txList := txService.GetTxList()
	assert.NotNil(t, txList)
	assert.Equal(t, 1, len((*txList).Transactions))
	assert.Equal(t, string(1), (*txList).Transactions[0].Asset.Nonce)
	assert.Equal(t, tx.SenderPubKey, (*txList).Transactions[0].SenderPubKey)
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

func TestGetTxListSecp256k1(t *testing.T) {
	var txService transaction.Service
	txService = transaction.TxService()
	for i := 1; i <= 10; i++ {
		tx := getTxUsingSecp256k1Account(i)
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

func getTxUsingSecp256k1Account(nonce int) transaction.Tx {
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

	return tx
}
