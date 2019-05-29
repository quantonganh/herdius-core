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

func TestInitBTC(t *testing.T) {
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
	eBalances["BTC"] = statedb.EBalance{Balance: uint64(0)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress1"

	es := &BTCSyncer{Account: account, ExBal: accountCache}
	es.ExtBalance = big.NewInt(1)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Balance, es.ExtBalance.Uint64(), "Balance should be updated with external balance")

}
func TestExternalBTCisGreater(t *testing.T) {
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
	eBalances["BTC"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress"

	es := &BTCSyncer{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(3)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(3), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["BTC"], big.NewInt(3), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["BTC"], true, "CurrentExtBalance ahould be updated")

	es.ExtBalance = big.NewInt(20)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.IsFirstEntry["BTC"], false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["BTC"], true, "CurrentExtBalance ahould be updated")

	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(20), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["BTC"], big.NewInt(20), "CurrentExtBalance ahould be updated")
}

func TestExternalBTCisLesser(t *testing.T) {
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
	eBalances["BTC"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress"

	es := &BTCSyncer{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Balance, es.ExtBalance.Uint64(), "should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(10), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["BTC"], big.NewInt(10), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["BTC"], true, "CurrentExtBalance should be updated")

	es.ExtBalance = big.NewInt(1)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)
	assert.Equal(t, cachedAcc.IsFirstEntry["BTC"], false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["BTC"], true, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(1), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["BTC"], big.NewInt(1), "CurrentExtBalance ahould be updated")
}

func TestBTC(t *testing.T) {
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
	eBalances["BTC"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress"

	accountCache.Set(account.Address, external.AccountCache{})

	es := &BTCSyncer{Account: account, ExBal: accountCache}
	es.ExtBalance = big.NewInt(10)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(10), "should be updated")

}
