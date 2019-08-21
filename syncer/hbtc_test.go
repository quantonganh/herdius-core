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

func TestInitHBTC(t *testing.T) {
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
	eBalances["HBTC"] = make(map[string]statedb.EBalance)
	eBalances["HBTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(0)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress1"
	account.FirstExternalAddress = map[string]string{"ETH": addr}

	hs := newHBTCSyncer()
	hs.syncer.Account = account
	hs.syncer.Storage = accountCache
	hs.syncer.ExtBalance[addr] = big.NewInt(1)
	hs.Update()
	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, hs.syncer.ExtBalance[addr].Uint64(), cachedAcc.Account.EBalances["HBTC"][addr].Balance, "HBTC Balance should be updated with external balance")
}

func TestExternalHBTCisGreater(t *testing.T) {
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
	eBalances["ETH"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}
	eBalances["HBTC"] = make(map[string]statedb.EBalance)
	eBalances["HBTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"
	account.FirstExternalAddress = map[string]string{"ETH": addr}

	hs := newHBTCSyncer()
	hs.syncer.Account = account
	hs.syncer.Storage = accountCache
	hs.syncer.ExtBalance[addr] = big.NewInt(1)

	hs.Update()

	hs.syncer.ExtBalance[addr] = big.NewInt(20)

	hs.Update()

	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, hs.syncer.ExtBalance[addr].Uint64(), cachedAcc.Account.EBalances["HBTC"][addr].Balance, "HBTC Balance should be updated")
}

func TestExternalHBTCisLesser(t *testing.T) {
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
	eBalances["ETH"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}
	eBalances["HBTC"] = make(map[string]statedb.EBalance)
	eBalances["HBTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"
	account.FirstExternalAddress = map[string]string{"ETH": addr}

	hs := newHBTCSyncer()
	hs.syncer.Account = account
	hs.syncer.Storage = accountCache
	hs.syncer.ExtBalance[addr] = big.NewInt(20)

	hs.Update()

	hs.syncer.ExtBalance[addr] = big.NewInt(10)

	hs.Update()

	cachedAcc, ok := accountCache.Get(account.Address)
	assert.Equal(t, ok, true, "cache should return account")
	assert.Equal(t, hs.syncer.ExtBalance[addr].Uint64(), cachedAcc.Account.EBalances["HBTC"][addr].Balance, "HBTC Balance should be updated")

}
