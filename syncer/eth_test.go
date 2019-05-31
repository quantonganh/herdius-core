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

	es := &EthSyncer{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(1)
	es.Nonce = 5
	es.BlockHeight = big.NewInt(2)

	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Balance should be updated with external balance")

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

	es := &EthSyncer{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(3)
	es.Nonce = 5
	es.BlockHeight = big.NewInt(2)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(3), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(3), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], true, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Balance should be updated with external balance")

	es.ExtBalance = big.NewInt(20)
	es.Nonce = 6
	es.BlockHeight = big.NewInt(3)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["ETH"], true, "CurrentExtBalance ahould be updated")

	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(20), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(20), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Balance should be updated with external balance")
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

	es := &EthSyncer{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Nonce = 6
	es.BlockHeight = big.NewInt(3)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(10), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(10), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], true, "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Balance should be updated with external balance")

	es.ExtBalance = big.NewInt(1)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["ETH"], true, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(1), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(1), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Balance should be updated with external balance")

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

	es := &EthSyncer{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)
	//ac := cachedAcc.UpdateLastExtBalanceByKey("testEthAddress", big.NewInt(10))

	//assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(10), "should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].LastBlockHeight, es.BlockHeight.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Nonce, es.Nonce, "Balance should be updated with external balance")

}
