package sync

import (
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/herdius/herdius-core/storage/db"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

func TestInit(t *testing.T) {
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
	eBalances["ETH"][addr] = statedb.EBalance{Address: addr, Balance: uint64(0)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress1"

	es := newEthSyncer()
	es.syncer.Account = account
	es.syncer.Storage = accountCache
	// Set external balance coming from infura
	es.syncer.ExtBalance[addr] = big.NewInt(1)
	es.syncer.Nonce[addr] = 5
	es.syncer.BlockHeight[addr] = big.NewInt(2)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].Balance, es.syncer.ExtBalance[addr].Uint64(), "Balance should be updated with external balance")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].LastBlockHeight, es.syncer.BlockHeight[addr].Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].Nonce, es.syncer.Nonce[addr], "Nonce should be updated with external Nonce")

}
func TestExternalIsGreater(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "ETH-1"
	addr2 := "ETH-2"
	storageKey := "ETH-" + addr
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["ETH"] = make(map[string]statedb.EBalance)
	eBalances["ETH"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}
	eBalances["ETH"][addr2] = statedb.EBalance{Address: addr2, Balance: uint64(18)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	es := newEthSyncer()
	es.syncer.Account = account
	es.syncer.Storage = accountCache
	// Set external balance coming from infura
	es.syncer.ExtBalance[addr] = big.NewInt(3)
	es.syncer.Nonce[addr] = 5
	es.syncer.BlockHeight[addr] = big.NewInt(2)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].Balance, es.syncer.ExtBalance[addr].Uint64(), "balance should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(3), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance[storageKey], big.NewInt(3), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry[storageKey], true, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].LastBlockHeight, es.syncer.BlockHeight[addr].Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].Nonce, es.syncer.Nonce[addr], "Nonce should be updated with external nonce")

	es.syncer.ExtBalance[addr] = big.NewInt(20)
	es.syncer.Nonce[addr] = 6
	es.syncer.BlockHeight[addr] = big.NewInt(3)
	es.Update()
	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].Balance, es.syncer.ExtBalance[addr].Uint64(), "Balance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry[storageKey], false, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate[storageKey], true, "IsNewAmountUpdate should be updated")

	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(20), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance[storageKey], big.NewInt(20), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].LastBlockHeight, es.syncer.BlockHeight[addr].Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].Nonce, es.syncer.Nonce[addr], "Nonce should be updated with external Nonce")
}

func TestExternalIsLesser(t *testing.T) {
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
	account.Address = "testEthAddress"

	bs := newBTCSyncer()
	bs.syncer.Account = account
	bs.syncer.Storage = accountCache
	// Set external balance coming from infura
	bs.syncer.ExtBalance[addr] = big.NewInt(10)
	bs.syncer.Nonce[addr] = 6
	bs.syncer.BlockHeight[addr] = big.NewInt(3)
	bs.Update()
	cachedAcc, ok := accountCache.Get(account.Address)

	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Balance, bs.syncer.ExtBalance[addr].Uint64(), "Balance should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(10), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance[storageKey], big.NewInt(10), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry[storageKey], true, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].LastBlockHeight, bs.syncer.BlockHeight[addr].Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Nonce, bs.syncer.Nonce[addr], "Nonce should be updated with external Nonce")

	bs.syncer.ExtBalance[addr] = big.NewInt(1)
	bs.syncer.Nonce[addr] = 7
	bs.syncer.BlockHeight[addr] = big.NewInt(4)
	bs.Update()
	cachedAcc, ok = accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.IsFirstEntry[storageKey], false, "IsFirstEntry should be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate[storageKey], true, "IsNewAmountUpdate should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(1), "LastExtBalance should be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance[storageKey], big.NewInt(1), "CurrentExtBalance should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].LastBlockHeight, bs.syncer.BlockHeight[addr].Uint64(), "BlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["BTC"][addr].Nonce, bs.syncer.Nonce[addr], "Nonce should be updated with external Nonce")
}

func TestCacheExistButWithoutAccountAsset(t *testing.T) {
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
	eBalances["ETH"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	accountCache.Set(account.Address, external.AccountCache{})

	es := newEthSyncer()
	es.syncer.Account = account
	es.syncer.Storage = accountCache
	// Set external balance coming from infura
	es.syncer.ExtBalance[addr] = big.NewInt(10)
	es.syncer.Nonce[addr] = 7
	es.syncer.BlockHeight[addr] = big.NewInt(4)
	es.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, cachedAcc.LastExtBalance[storageKey], big.NewInt(10), "should be updated")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].LastBlockHeight, es.syncer.BlockHeight[addr].Uint64(), "LastBlockHeight should be updated with external height")
	assert.Equal(t, cachedAcc.Account.EBalances["ETH"][addr].Nonce, es.syncer.Nonce[addr], "Nonce should be updated with external Nonce")

}
