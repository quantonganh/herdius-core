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
func TestHERShouldNOTChangeOtherASSET(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "ETH-1"
	storageKey := "ETH-" + addr
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["ETH"] = make(map[string]statedb.EBalance)
	eBalances["ETH"][addr] = statedb.EBalance{Balance: uint64(1)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	es := &HERToken{Account: account, Storage: accountCache}
	leb := make(map[string]*big.Int)
	leb[storageKey] = big.NewInt(9)
	accountCache.Set(account.Address, external.AccountCache{LastExtBalance: leb})

	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(1)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")

	es.Update()
	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")

	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].Balance, uint64(1), "Balance should not be updated with external balance")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(9), "LastExtBalance should not be updated with external balance")
	assert.Equal(t, cachedAcc.Account.Balance, big.NewInt(1).Uint64(), "Balance should not be updated with external balance")

}

func TestHERExternalETHisGreater(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "ETH-1"
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["ETH"] = make(map[string]statedb.EBalance)
	eBalances["ETH"][addr] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress5"

	es := &HERToken{Account: account, Storage: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(3)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)

	last, ispresent := es.Storage.Get(es.Account.Address)
	assert.Equal(t, ispresent, false, "cache should not return account")

	accountCache.Set(account.Address, last)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")

	assert.Equal(t, cachedAcc.Account.Balance, es.ExtBalance.Uint64(), "ExtBalance should be updated")
	assert.Equal(t, cachedAcc.LastExtHERBalance, big.NewInt(3), "LastExtHERBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtHERBalance, big.NewInt(3), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstHEREntry, true, "IsFirstHEREntry should be updated")

	es.ExtBalance = big.NewInt(10)
	es.Update()
	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.IsFirstHEREntry, false, "IsFirstHEREntry should be updated")
	assert.Equal(t, cachedAcc.IsNewHERAmountUpdate, true, "IsNewHERAmountUpdate should be updated")

	assert.Equal(t, cachedAcc.Account.Balance, es.ExtBalance.Uint64(), "Balance should be updated")
	assert.Equal(t, cachedAcc.LastExtHERBalance, big.NewInt(10), "LastExtHERBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtHERBalance, big.NewInt(10), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.ExternalNonce, uint64(7), "ExternalNonce should be updated")

}

func TestHERExternalETHisLesser(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "ETH-1"
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["ETH"] = make(map[string]statedb.EBalance)
	eBalances["ETH"][addr] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress00"

	es := &HERToken{Account: account, Storage: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)

	last, ispresent := es.Storage.Get(es.Account.Address)
	assert.Equal(t, ispresent, false, "cache not should return account")

	accountCache.Set(account.Address, last)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")

	assert.Equal(t, cachedAcc.Account.Balance, es.ExtBalance.Uint64(), "Balance should be updated")
	assert.Equal(t, cachedAcc.LastExtHERBalance, big.NewInt(10), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtHERBalance, big.NewInt(10), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstHEREntry, true, "IsFirstHEREntry should be updated")
	assert.Equal(t, cachedAcc.Account.ExternalNonce, uint64(7), "ExternalNonce should be updated with ecternal nonce")
	assert.Equal(t, cachedAcc.Account.LastBlockHeight, uint64(4), "LastBlockHeight should be updated with external height")

	es.ExtBalance = big.NewInt(1)
	es.Update()
	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.Balance, es.ExtBalance.Uint64(), "Balance should be updated ")
	assert.Equal(t, cachedAcc.LastExtHERBalance, big.NewInt(1), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtHERBalance, big.NewInt(1), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstHEREntry, false, "IsFirstHEREntry should be updated")
	assert.Equal(t, cachedAcc.Account.ExternalNonce, uint64(7), "ExternalNonce should be updated with external nonce")
	assert.Equal(t, cachedAcc.Account.LastBlockHeight, uint64(4), "LastBlockHeight should be updated with external height")

}
