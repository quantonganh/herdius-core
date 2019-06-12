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
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "BTC-1"
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["BTC"] = make(map[string]statedb.EBalance)
	eBalances["BTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(0)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress1"

	es := newBTCSyncer()
	es.Account = account
	es.Storage = accountCache
	es.ExtBalance[addr] = big.NewInt(1)
	es.Nonce[addr] = 7
	es.BlockHeight[addr] = big.NewInt(4)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Balance, es.ExtBalance[addr].Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].LastBlockHeight, es.BlockHeight[addr].Uint64(), "block Height should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Nonce, es.Nonce[addr], "Nonce should be updated with external nonce")
}
func TestExternalBTCisGreater(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "BTC-1"
	storageKey := "BTC-" + addr
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["BTC"] = make(map[string]statedb.EBalance)
	eBalances["BTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress"

	es := newBTCSyncer()
	es.Account = account
	es.Storage = accountCache
	es.ExtBalance[addr] = big.NewInt(3)
	es.Nonce[addr] = 5
	es.BlockHeight[addr] = big.NewInt(2)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Balance, es.ExtBalance[addr].Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(3), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance[storageKey], big.NewInt(3), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry[storageKey], true, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].LastBlockHeight, es.BlockHeight[addr].Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Nonce, es.Nonce[addr], "Nonce should be updated with external nonce")

	es.ExtBalance[addr] = big.NewInt(20)
	es.Nonce[addr] = 6
	es.BlockHeight[addr] = big.NewInt(3)
	es.Update()

	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Balance, es.ExtBalance[addr].Uint64(), "Balance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry[storageKey], false, "LastBlockHeight should be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate[storageKey], true, "IsNewAmountUpdate should be updated")

	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(20), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance[storageKey], big.NewInt(20), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].LastBlockHeight, es.BlockHeight[addr].Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Nonce, es.Nonce[addr], "Nonce should be updated with external nonce")

}

func TestExternalBTCisLesser(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "BTC-1"
	storageKey := "BTC-" + addr
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["BTC"] = make(map[string]statedb.EBalance)
	eBalances["BTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress"

	es := newBTCSyncer()
	es.Account = account
	es.Storage = accountCache
	// Set external balance coming from infura
	es.ExtBalance[addr] = big.NewInt(10)
	es.Nonce[addr] = 6
	es.BlockHeight[addr] = big.NewInt(3)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Balance, es.ExtBalance[addr].Uint64(), " Balance should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(10), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance[storageKey], big.NewInt(10), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry[storageKey], true, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].LastBlockHeight, es.BlockHeight[addr].Uint64(), "BlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Nonce, es.Nonce[addr], "Nonce should be updated with external nonce")

	es.ExtBalance[addr] = big.NewInt(1)
	es.Nonce[addr] = 7
	es.BlockHeight[addr] = big.NewInt(4)
	es.Update()
	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.IsFirstEntry[storageKey], false, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate[storageKey], true, "IsNewAmountUpdate should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(1), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance[storageKey], big.NewInt(1), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].LastBlockHeight, es.BlockHeight[addr].Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Nonce, es.Nonce[addr], "Nonce should be updated with external nonce")

}

func TestBTC(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "BTC-1"
	storageKey := "BTC-" + addr
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["BTC"] = make(map[string]statedb.EBalance)
	eBalances["BTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress"

	accountCache.Set(account.Address, external.AccountCache{})

	es := newBTCSyncer()
	es.Account = account
	es.Storage = accountCache
	es.ExtBalance[addr] = big.NewInt(10)
	es.Nonce[addr] = 7
	es.BlockHeight[addr] = big.NewInt(4)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(10), "should be updated")

}

func TestNoResponseFromAPI(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "BTC-1"
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["BTC"] = make(map[string]statedb.EBalance)
	eBalances["BTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress"

	es := newBTCSyncer()
	es.Account = account
	es.Storage = accountCache
	es.ExtBalance[addr] = big.NewInt(1)
	es.Nonce[addr] = 0
	es.BlockHeight[addr] = big.NewInt(0)
	es.Update()

	es.ExtBalance[addr] = nil
	assert.Panics(t, es.Update, "")

}
