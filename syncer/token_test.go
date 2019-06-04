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
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(1)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	es := &HERToken{Account: account, ExBal: accountCache}
	leb := make(map[string]*big.Int)
	leb["ETH"] = big.NewInt(9)
	accountCache.Set(account.Address, external.AccountCache{LastExtBalance: leb})

	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(1)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	cachedAcc, _ := accountCache.Get(account.Address)

	//cachedAcc.UpdateCurrentExtHERBalance(big.NewInt(1))

	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, uint64(1), "Balance should not be updated with external balance")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(9), "Balance should not be updated with external balance")

	assert.Equal(t, cachedAcc.Account.Balance, big.NewInt(1).Uint64(), "Balance should not be updated with external balance")

}

func TestHERExternalETHisGreater(t *testing.T) {
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
	account.Address = "testEthAddress5"

	es := &HERToken{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(3)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)

	last, _ := es.ExBal.Get(es.Account.Address)

	accountCache.Set(account.Address, last)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtHERBalance, big.NewInt(3), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtHERBalance, big.NewInt(3), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstHEREntry, true, "CurrentExtBalance ahould be updated")

	es.ExtBalance = big.NewInt(10)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.IsFirstHEREntry, false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsNewHERAmountUpdate, true, "CurrentExtBalance ahould be updated")

	assert.Equal(t, cachedAcc.Account.Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtHERBalance, big.NewInt(10), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtHERBalance, big.NewInt(10), "CurrentExtBalance ahould be updated")
}

func TestHERExternalETHisLesser(t *testing.T) {
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
	account.Address = "testEthAddress00"

	es := &HERToken{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)

	last, _ := es.ExBal.Get(es.Account.Address)

	accountCache.Set(account.Address, last)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtHERBalance, big.NewInt(10), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtHERBalance, big.NewInt(10), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstHEREntry, true, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.Account.Nonce, uint64(7), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.Account.LastBlockHeight, uint64(4), "Total fetched transactions should be 20")

	es.ExtBalance = big.NewInt(1)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtHERBalance, big.NewInt(1), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtHERBalance, big.NewInt(1), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstHEREntry, false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.Account.Nonce, uint64(7), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.Account.LastBlockHeight, uint64(4), "Total fetched transactions should be 20")

}
