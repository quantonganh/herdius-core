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

	es := &BTCSyncer{Account: account, Storage: accountCache}
	es.ExtBalance = big.NewInt(1)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Balance, es.ExtBalance.Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].LastBlockHeight, es.BlockHeight.Uint64(), "block Height should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Nonce, es.Nonce, "Nonce should be updated with external nonce")
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

	es := &BTCSyncer{Account: account, Storage: accountCache}
	es.ExtBalance = big.NewInt(3)
	es.Nonce = 5
	es.BlockHeight = big.NewInt(2)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(3), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["BTC"], big.NewInt(3), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["BTC"], true, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].LastBlockHeight, es.BlockHeight.Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Nonce, es.Nonce, "Nonce should be updated with external nonce")

	es.ExtBalance = big.NewInt(20)
	es.Nonce = 6
	es.BlockHeight = big.NewInt(3)
	es.Update()

	cachedAcc, _ = accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Balance, es.ExtBalance.Uint64(), "Balance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["BTC"], false, "LastBlockHeight should be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["BTC"], true, "IsNewAmountUpdate should be updated")

	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(20), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["BTC"], big.NewInt(20), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].LastBlockHeight, es.BlockHeight.Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Nonce, es.Nonce, "Nonce should be updated with external nonce")

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

	es := &BTCSyncer{Account: account, Storage: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Nonce = 6
	es.BlockHeight = big.NewInt(3)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Balance, es.ExtBalance.Uint64(), " Balance should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(10), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["BTC"], big.NewInt(10), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["BTC"], true, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].LastBlockHeight, es.BlockHeight.Uint64(), "BlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Nonce, es.Nonce, "Nonce should be updated with external nonce")

	es.ExtBalance = big.NewInt(1)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)
	assert.Equal(t, cachedAcc.IsFirstEntry["BTC"], false, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["BTC"], true, "IsNewAmountUpdate should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(1), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["BTC"], big.NewInt(1), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].LastBlockHeight, es.BlockHeight.Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"].Nonce, es.Nonce, "Nonce should be updated with external nonce")

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

	es := &BTCSyncer{Account: account, Storage: accountCache}
	es.ExtBalance = big.NewInt(10)
	es.Nonce = 7
	es.BlockHeight = big.NewInt(4)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.LastExtBalance["BTC"], big.NewInt(10), "should be updated")

}

func TestNoResponseFromAPI(t *testing.T) {
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

	es := &BTCSyncer{Account: account, Storage: accountCache}
	es.ExtBalance = big.NewInt(1)
	es.Nonce = 0
	es.BlockHeight = big.NewInt(0)
	es.Update()

	es.ExtBalance = nil
	//es.Update()
	assert.Panics(t, es.Update, "")

}
