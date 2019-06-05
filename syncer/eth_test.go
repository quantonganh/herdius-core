package sync

import (
	"math/big"
	"os"
	"testing"

	"github.com/herdius/herdius-core/storage/db"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/stretchr/testify/assert"
)

/*

1) test if external balance of ETH is getting updated first time
2) test if external ETH is greater than already existing eth
*/

func TestInit(t *testing.T) {
	var (
		accountCache external.BalanceStorage
		eBalances    map[string]statedb.EBalance
	)
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(0)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress1"

	es := &EthSyncer{Account: account, Storage: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(1)
	es.Nonce = 5
	es.BlockHeight = big.NewInt(2)

	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t,ok,true,"cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Nonce should be updated with external Nonce")

}
func TestExternalETHisGreater(t *testing.T) {
	var (
		accountCache external.BalanceStorage
		eBalances    map[string]statedb.EBalance
	)
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	es := &EthSyncer{Account: account, Storage: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(3)
	es.Nonce = 5
	es.BlockHeight = big.NewInt(2)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t,ok,true,"cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "balance should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(3), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(3), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], true, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Nonce should be updated with external nonce")

	es.ExtBalance = big.NewInt(20)
	es.Nonce = 6
	es.BlockHeight = big.NewInt(3)
	es.Update()
	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t,ok,true,"cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Balance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], false, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["ETH"], true, "IsNewAmountUpdate should be updated")

	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(20), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(20), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Nonce should be updated with external Nonce")
}

func TestExternalETHisLesser(t *testing.T) {
	var (
		accountCache external.BalanceStorage
		eBalances    map[string]statedb.EBalance
	)
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	es := &EthSyncer{Account: account, Storage: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Nonce = 6
	es.BlockHeight = big.NewInt(3)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)

	assert.Equal(t,ok,true,"cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Balance should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(10), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(10), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], true, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Nonce should be updated with external Nonce")

	es.ExtBalance = big.NewInt(1)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	es.Update()
	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t,ok,true,"cache should return account")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], false, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["ETH"], true, "IsNewAmountUpdate should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(1), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(1), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "BlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Nonce should be updated with external Nonce")

}

// Test if cache exists but Account cache dont have eth asset
func Test(t *testing.T) {
	var (
		accountCache external.BalanceStorage
		eBalances    map[string]statedb.EBalance
	)
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	accountCache.Set(account.Address, external.AccountCache{})

	es := &EthSyncer{Account: account, Storage: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t,ok,true,"cache should return account")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(10), "should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Nonce should be updated with external Nonce")

}
