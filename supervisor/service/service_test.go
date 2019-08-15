package service

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"testing"

	"github.com/herdius/herdius-core/blockchain/protobuf"
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
		Address:              "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		FirstExternalAddress: make(map[string]string),
	}
	account = updateAccount(account, tx)
	assert.True(t, len(account.EBalances) > 0)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances[symbol][extSenderAddress].Address)
	assert.Equal(t, extSenderAddress, account.FirstExternalAddress[symbol])
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
		Address:              "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		FirstExternalAddress: make(map[string]string),
	}
	account = updateAccount(account, tx)
	assert.True(t, len(account.EBalances) == 1)
	assert.True(t, len(account.EBalances[symbol]) == 1)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances[symbol][extSenderAddress].Address)
	assert.Equal(t, extSenderAddress, account.FirstExternalAddress[symbol])

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
	assert.Equal(t, newExtSenderAddress, account.FirstExternalAddress[newSymbol])
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
		Address:              "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		FirstExternalAddress: make(map[string]string),
	}
	account = updateAccount(account, tx)
	assert.True(t, len(account.EBalances) > 0)
	assert.Equal(t, extSenderAddress, account.FirstExternalAddress[symbol])
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.EBalances[symbol][extSenderAddress].Address)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.FirstExternalAddress[symbol])

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
	assert.Equal(t, uint64(0), account.EBalances[symbol][extSenderAddress].Balance)
	assert.Equal(t, tx.Asset.ExternalSenderAddress, account.FirstExternalAddress[symbol])

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
	var txService transaction.Service = transaction.TxService()
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
	var txService transaction.Service = transaction.TxService()
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
	lastBlock := &protobuf.BaseBlock{}
	supsvc := &Supervisor{}
	supsvc.SetWriteMutex()
	supsvc.AddValidator([]byte{1}, "add-01")
	supsvc.AddValidator([]byte{1}, "add-02")
	supsvc.SetWriteMutex()
	txs := &txbyte.Txs{}
	_, err := supsvc.ShardToValidators(lastBlock, txs, nil, nil)
	assert.Error(t, err)
}

func TestShardToValidatorsTrue(t *testing.T) {
	lastBlock := &protobuf.BaseBlock{}
	supsvc := &Supervisor{}
	supsvc.SetWriteMutex()
	supsvc.AddValidator([]byte{1}, "add-01")
	supsvc.AddValidator([]byte{1}, "add-02")
	supsvc.SetWriteMutex()
	txs := &txbyte.Txs{}
	dir, err := ioutil.TempDir("", "yeezy")
	assert.NoError(t, err)
	trie = statedb.GetState(dir)
	root, err := trie.Commit(nil)
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	_, err = supsvc.ShardToValidators(lastBlock, txs, nil, root)
	assert.NoError(t, err)
}

func TestUpdateStateWithNewExternalBalance(t *testing.T) {
	dir, err := ioutil.TempDir("", "temp-dir")
	assert.NoError(t, err)

	extAddr := "external-address-01"
	asset := "external-asset"
	storageKey := asset + "-" + extAddr
	eBalance := statedb.EBalance{
		Address: extAddr,
		Balance: 0,
	}
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances[asset] = make(map[string]statedb.EBalance)
	eBalances[asset][extAddr] = eBalance
	herAccount := statedb.Account{
		Address:   "her-address-01",
		EBalances: eBalances,
	}

	assert.Equal(t, herAccount.EBalances[asset][extAddr].Balance, uint64(0))
	trie = statedb.GetState(dir)
	sactbz, err := cdc.MarshalJSON(herAccount)
	assert.NoError(t, err)
	err = trie.TryUpdate([]byte(herAccount.Address), sactbz)
	assert.NoError(t, err)

	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountStorage = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()
	currentExternalBal := make(map[string]*big.Int)
	currentExternalBal[asset] = big.NewInt(int64(math.Pow10(18)))

	lastExternalBal := make(map[string]*big.Int)
	lastExternalBal[asset] = big.NewInt(int64(0))

	isFirstEntry := make(map[string]bool)
	isFirstEntry[asset] = true

	eBalance.Balance = uint64(math.Pow10(18))
	eBalances[asset][extAddr] = eBalance
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
	assert.False(t, res.IsFirstEntry[storageKey])

	defer os.RemoveAll(dir)
}

func TestUpdateAccountLockedBalance(t *testing.T) {
	symbol := "ETH"
	lockedAmount := uint64(10)
	extSenderAddress := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	asset := &pluginproto.Asset{
		Symbol:                symbol,
		ExternalSenderAddress: extSenderAddress,
		Nonce:                 1,
		Network:               "Herdius",
		LockedAmount:          lockedAmount,
	}
	tx := &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "lock",
	}

	extAddr := "0xD8f647855876549d2623f52126CE40D053a2ef6A"

	eBalance := statedb.EBalance{
		Address: extAddr,
		Balance: 0,
	}
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances[symbol] = make(map[string]statedb.EBalance)
	eBalances[symbol][extAddr] = eBalance

	account := &statedb.Account{
		Address:              "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		FirstExternalAddress: make(map[string]string),
		EBalances:            eBalances,
	}
	account = updateAccountLockedBalance(account, tx)
	assert.Equal(t, lockedAmount, account.LockedBalance[symbol][extSenderAddress])
}
func TestUpdateAccountLockedBalanceMintHBTCFirst(t *testing.T) {
	symbol := "BTC"
	lockedAmount := uint64(10)
	extSenderAddress := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	asset := &pluginproto.Asset{
		Symbol:                symbol,
		ExternalSenderAddress: extSenderAddress,
		Nonce:                 1,
		Network:               "Herdius",
		LockedAmount:          lockedAmount,
		Value:                 1,
	}
	tx := &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "lock",
	}

	extAddr := "0xD8f647855876549d2623f52126CE40D053a2ef6A"

	eBalance := statedb.EBalance{
		Address: extAddr,
		Balance: 0,
	}
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances[symbol] = make(map[string]statedb.EBalance)
	eBalances[symbol][extAddr] = eBalance

	account := &statedb.Account{
		Address:              "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		FirstExternalAddress: make(map[string]string),
		EBalances:            eBalances,
	}
	account.FirstExternalAddress["ETH"] = "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	account = updateAccountLockedBalance(account, tx)
	assert.Equal(t, lockedAmount, account.LockedBalance[symbol][extSenderAddress])
	assert.Equal(t, uint64(0), account.EBalances["HBTC"]["0xD8f647855876549d2623f52126CE40D053a2ef6A"].Balance)

}

func TestUpdateRedeemAccountLockedBalance(t *testing.T) {
	symbol := "ETH"
	lockedAmount := uint64(10)
	extSenderAddress := "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	asset := &pluginproto.Asset{
		Symbol:                symbol,
		ExternalSenderAddress: extSenderAddress,
		Nonce:                 1,
		Network:               "Herdius",
		LockedAmount:          lockedAmount,
	}
	tx := &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "lock",
	}

	extAddr := "0xD8f647855876549d2623f52126CE40D053a2ef6A"

	eBalance := statedb.EBalance{
		Address: extAddr,
		Balance: 0,
	}
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances[symbol] = make(map[string]statedb.EBalance)
	eBalances[symbol][extAddr] = eBalance
	account := &statedb.Account{
		Address:              "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		FirstExternalAddress: make(map[string]string),
		EBalances:            eBalances,
	}
	account = updateAccountLockedBalance(account, tx)
	assert.Equal(t, lockedAmount, account.LockedBalance[symbol][extSenderAddress])

	// Redeem test
	symbol = "ETH"
	value := uint64(5)
	extSenderAddress = "0xD8f647855876549d2623f52126CE40D053a2ef6A"
	asset = &pluginproto.Asset{
		Symbol:                symbol,
		ExternalSenderAddress: extSenderAddress,
		Nonce:                 2,
		Network:               "Herdius",
		RedeemedAmount:        value,
	}
	tx = &pluginproto.Tx{
		SenderAddress: "HHy1CuT3UxCGJ3BHydLEvR5ut6HRy2qUvm",
		Asset:         asset,
		Type:          "redeem",
	}

	account = updateRedeemAccountLockedBalance(account, tx)
	assert.Equal(t, value, account.LockedBalance[symbol][extSenderAddress])
}

func TestValidatorGroups(t *testing.T) {
	tests := []struct {
		name                             string
		numValidators                    int
		desiredNumGroups                 int
		expectedNumGroups                int
		expectedNumValidatorsInEachGroup int
		expectedNumValidatorsInLastGroup int
	}{
		{"validator less than group", 1, 2, 1, 1, 1},
		{"validator equal group", 2, 2, 2, 1, 1},
		{"validator greater than group", 500, 5, 5, 100, 100},
		{"validator not divisible to each group #1", 501, 5, 5, 101, 97},
		{"validator not divisible to each group #2", 499, 5, 5, 100, 99},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			supsvc := &Supervisor{}
			supsvc.SetWriteMutex()
			for i := 1; i <= tc.numValidators; i++ {
				supsvc.AddValidator([]byte{1}, fmt.Sprintf("validator-%d", i))
			}
			groups := supsvc.validatorGroups(tc.desiredNumGroups)
			if len(groups) != tc.expectedNumGroups {
				t.Errorf("Unexpected number of groups, want: %d, got: %d", tc.expectedNumGroups, len(groups))
			}

			for i := 0; i < len(groups)-1; i++ {
				if len(groups[i]) != tc.expectedNumValidatorsInEachGroup {
					t.Errorf("Unexpected number of validators in each group, want: %d, got: %d", tc.expectedNumValidatorsInEachGroup, len(groups[i]))
				}
			}
			if len(groups[len(groups)-1]) != tc.expectedNumValidatorsInLastGroup {
				t.Errorf("Unexpected number of validators in last group, want: %d, got: %d", tc.expectedNumValidatorsInLastGroup, len(groups[len(groups)-1]))
			}
		})
	}
}

func TestTxsGroups(t *testing.T) {
	tests := []struct {
		name              string
		numTxs            int
		desiredNumGroups  int
		expectedNumGroups int
	}{
		{"txs less than group", 1, 2, 1},
		{"txs equal group", 2, 2, 2},
		{"txs greater than group", 4500, 5, 5},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			supsvc := &Supervisor{}
			txs := make([]*transaction.Tx, tc.numTxs)
			txList := &transaction.TxList{Transactions: txs}

			groups := supsvc.txsGroups(txList, tc.desiredNumGroups)
			if len(groups) != tc.expectedNumGroups {
				t.Errorf("Unexpected number of groups, want: %d, got: %d", tc.expectedNumGroups, len(groups))
			}
		})
	}
}
