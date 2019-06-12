package service

import (
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"testing"

	"github.com/herdius/herdius-core/storage/db"
	"github.com/herdius/herdius-core/storage/state/statedb"

	ed25519 "github.com/herdius/herdius-core/crypto/ed"
	pluginproto "github.com/herdius/herdius-core/hbi/protobuf"

	"github.com/herdius/herdius-core/crypto/secp256k1"
	"github.com/herdius/herdius-core/supervisor/transaction"

	external "github.com/herdius/herdius-core/storage/exbalance"

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
	symbol := "ETH"
	extSenderAddress := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	asset := &pluginproto.Asset{
		Symbol:                symbol,
		ExternalSenderAddress: extSenderAddress,
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
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances[symbol][extSenderAddress].Address)
}

func TestRegisterMultipleExternalAssets(t *testing.T) {
	symbol := "ETH"
	extSenderAddress := " 0xD8f647855876549d2623f52126CE40D053a2ef6A"
	// First add ETH
	asset := &pluginproto.Asset{
		Symbol:                symbol,
		ExternalSenderAddress: extSenderAddress,
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
	assert.True(t, len(account.EBalances) == 1)
	assert.True(t, len(account.EBalances[symbol]) == 1)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances[symbol][extSenderAddress].Address)

	newSymbol := "BTC"
	newExtSenderAddress := "Bitcoin-Address"
	// Second add BTC
	asset = &pluginproto.Asset{
		Symbol:                newSymbol,
		ExternalSenderAddress: newExtSenderAddress,
		Nonce:                 2,
		Network:               "Herdius",
	}
	tx = &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "update",
	}

	account = updateAccount(account, tx)
	assert.True(t, len(account.EBalances) == 2)
	assert.True(t, len(account.EBalances[newSymbol]) == 1)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances[newSymbol][newExtSenderAddress].Address)
}

func TestUpdateExternalAccountBalance(t *testing.T) {
	symbol := "ETH"
	extSenderAddress := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	asset := &pluginproto.Asset{
		Symbol:                symbol,
		ExternalSenderAddress: extSenderAddress,
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
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances[symbol][extSenderAddress].Address)

	asset = &pluginproto.Asset{
		Symbol:                symbol,
		ExternalSenderAddress: extSenderAddress,
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
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances[symbol][extSenderAddress].Address)
	assert.Equal(t, uint64(15), account.EBalances[symbol][extSenderAddress].Balance)
}

func TestIsExternalAssetAddressExistTrue(t *testing.T) {
	addr := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	eBal := statedb.EBalance{Address: addr}
	eBals := make(map[string]map[string]statedb.EBalance)
	eBals["ETH"] = make(map[string]statedb.EBalance)
	eBals["ETH"][addr] = eBal
	account := &statedb.Account{
		Address:   "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		EBalances: eBals,
	}
	assert.True(t, isExternalAssetAddressExist(account, "ETH", addr))
}
func TestIsExternalAssetAddressExistFalse(t *testing.T) {
	addr := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	eBals := make(map[string]map[string]statedb.EBalance)
	account := &statedb.Account{
		Address:   "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		EBalances: eBals,
	}
	assert.False(t, isExternalAssetAddressExist(account, "ETH", addr))
}

func TestExternalAssetWithdrawFromAnAccount(t *testing.T) {
	addr := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	eBal := statedb.EBalance{Balance: 10, Address: addr}
	eBals := make(map[string]map[string]statedb.EBalance)
	eBals["ETH"] = make(map[string]statedb.EBalance)
	eBals["ETH"][addr] = eBal
	account := &statedb.Account{
		Address:   "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		EBalances: eBals,
	}
	withdraw(account, "ETH", addr, 5)
	assert.Equal(t, uint64(5), account.EBalances["ETH"][addr].Balance)
}

func TestExternalAssetDepositToAnAccount(t *testing.T) {
	addr := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	eBal := statedb.EBalance{Balance: 10, Address: addr}
	eBals := make(map[string]map[string]statedb.EBalance)
	eBals["ETH"] = make(map[string]statedb.EBalance)
	eBals["ETH"][addr] = eBal
	account := &statedb.Account{
		Address:   "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		EBalances: eBals,
	}
	deposit(account, "ETH", addr, 5)
	assert.Equal(t, uint64(15), account.EBalances["ETH"][addr].Balance)
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

func TestShardToValidatorsFalse(t *testing.T) {
	supsvc := &Supervisor{}
	supsvc.AddValidator([]byte{1}, "add-01")
	supsvc.AddValidator([]byte{1}, "add-02")
	supsvc.SetWriteMutex()
	txs := &txbyte.Txs{}
	err := supsvc.ShardToValidators(txs, nil, nil)
	assert.Error(t, err)
}

func TestShardToValidatorsTrue(t *testing.T) {
	supsvc := &Supervisor{}
	supsvc.AddValidator([]byte{1}, "add-01")
	supsvc.AddValidator([]byte{1}, "add-02")
	supsvc.SetWriteMutex()
	txs := &txbyte.Txs{}
	dir, err := ioutil.TempDir("", "yeezy")
	trie = statedb.GetState(dir)
	root, err := trie.Commit(nil)
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	err = supsvc.ShardToValidators(txs, nil, root)
	assert.NoError(t, err)
}

func TestUpdateStateWithNewExternalBalance(t *testing.T) {
	dir, err := ioutil.TempDir("", "temp-dir")

	extAddr := "external-address-01"
	eBalance := statedb.EBalance{
		Address: extAddr,
		Balance: 0,
	}
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["external-asset"] = make(map[string]statedb.EBalance)
	eBalances["external-asset"][extAddr] = eBalance
	herAccount := statedb.Account{
		Address:   "her-address-01",
		EBalances: eBalances,
	}

	assert.Equal(t, herAccount.EBalances["external-asset"][extAddr].Balance, uint64(0))
	trie = statedb.GetState(dir)
	sactbz, err := cdc.MarshalJSON(herAccount)
	err = trie.TryUpdate([]byte(herAccount.Address), sactbz)

	assert.NoError(t, err)

	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountStorage = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()
	currentExternalBal := make(map[string]*big.Int)
	currentExternalBal["external-asset"] = big.NewInt(int64(math.Pow10(18)))

	lastExternalBal := make(map[string]*big.Int)
	lastExternalBal["external-asset"] = big.NewInt(int64(0))

	isFirstEntry := make(map[string]bool)
	isFirstEntry["external-asset"] = true

	eBalance.Balance = uint64(math.Pow10(18))
	eBalances["external-asset"][extAddr] = eBalance
	herAccount.EBalances = eBalances
	herCacheAccount := external.AccountCache{
		Account:           herAccount,
		CurrentExtBalance: currentExternalBal,
		LastExtBalance:    lastExternalBal,
		IsFirstEntry:      isFirstEntry,
	}
	accountStorage.Set(extAddr, herCacheAccount)

	updateStateWithNewExternalBalance(trie)

	res, ok := accountStorage.Get(extAddr)
	assert.True(t, ok)
	assert.False(t, res.IsFirstEntry["external-asset"])

	defer os.RemoveAll(dir)
}
