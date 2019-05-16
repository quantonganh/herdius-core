package service

import (
	"testing"

	"github.com/herdius/herdius-core/storage/state/statedb"

	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	pluginproto "github.com/herdius/herdius-core/hbi/protobuf"

	"github.com/herdius/herdius-core/crypto/secp256k1"
	"github.com/herdius/herdius-core/supervisor/transaction"
	txbyte "github.com/herdius/herdius-core/tx"
	"github.com/stretchr/testify/assert"
)

func TestRegisterNewHERAddress(t *testing.T) {
	asset := &pluginproto.Asset{
		Symbol: "HER",
	}
	tx := &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "update",
	}
	account := &statedb.Account{}
	account = updateAccount(account, tx)
	assert.Equal(t, tx.SenderAddress, account.Address)
}

func TestUpdateHERAccountBalance(t *testing.T) {
	asset := &pluginproto.Asset{
		Symbol: "HER",
	}
	tx := &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "update",
	}
	account := &statedb.Account{}
	account = updateAccount(account, tx)
	assert.Equal(t, tx.SenderAddress, account.Address)
	assert.Equal(t, account.Balance, uint64(0))

	// Update 10 HER tokens to existing HER Account
	asset = &pluginproto.Asset{
		Symbol: "HER",
		Value:  10,
		Nonce:  2,
	}
	tx = &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "update",
	}
	account = updateAccount(account, tx)
	assert.Equal(t, tx.SenderAddress, account.Address)
	assert.Equal(t, account.Balance, uint64(10))
	assert.Equal(t, account.Nonce, uint64(2))
}

func TestRegisterNewETHAddress(t *testing.T) {
	asset := &pluginproto.Asset{
		Symbol:                "ETH",
		ExternalSenderAddress: "0xD8f647855876549d2623f52126CE40D053a2ef6A",
		Nonce:                 1,
		Network:               "Herdius",
	}
	tx := &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "update",
	}
	account := &statedb.Account{
		Address: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
	}
	account = updateAccount(account, tx)
	assert.True(t, len(account.EBalances) > 0)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances["ETH"].Address)
}

func TestUpdateExternalAccountBalance(t *testing.T) {
	asset := &pluginproto.Asset{
		Symbol:                "ETH",
		ExternalSenderAddress: "0xD8f647855876549d2623f52126CE40D053a2ef6A",
		Nonce:                 1,
		Network:               "Herdius",
	}
	tx := &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "update",
	}
	account := &statedb.Account{
		Address: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
	}
	account = updateAccount(account, tx)
	assert.True(t, len(account.EBalances) > 0)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances["ETH"].Address)

	asset = &pluginproto.Asset{
		Symbol:                "ETH",
		ExternalSenderAddress: "0xD8f647855876549d2623f52126CE40D053a2ef6A",
		Nonce:                 2,
		Network:               "Herdius",
		Value:                 15,
	}
	tx = &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "update",
	}

	account = updateAccount(account, tx)
	assert.True(t, len(account.EBalances) > 0)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances["ETH"].Address)
	assert.Equal(t, uint64(15), account.EBalances["ETH"].Balance)
}

func TestIsExternalAssetAddressAvailableTrue(t *testing.T) {
	eBal := statedb.EBalance{
		Address: "0xD8f647855876549d2623f52126CE40D053a2ef6A",
	}
	eBals := make(map[string]statedb.EBalance)
	eBals["ETH"] = eBal
	account := &statedb.Account{
		Address:   "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		EBalances: eBals,
	}
	assert.True(t, isExternalAssetAddressAvailable(account, "ETH"))
}
func TestIsExternalAssetAddressAvailableFalse(t *testing.T) {
	eBal := statedb.EBalance{}
	eBals := make(map[string]statedb.EBalance)
	eBals["ETH"] = eBal
	account := &statedb.Account{
		Address:   "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		EBalances: eBals,
	}
	assert.False(t, isExternalAssetAddressAvailable(account, "ETH"))
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
	return tx
}

func TestShardToValidators(t *testing.T) {
	supsvc := &Supervisor{}
	supsvc.AddValidator([]byte{1}, "add-01")
	supsvc.AddValidator([]byte{1}, "add-02")
	supsvc.SetWriteMutex()
	txs := &txbyte.Txs{}
	err := supsvc.ShardToValidators(txs, nil, nil)
	assert.Nil(t, err)
}
