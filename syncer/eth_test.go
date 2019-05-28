package sync

import (
	"math/big"
	"testing"

	"github.com/herdius/herdius-core/storage/cache"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/stretchr/testify/assert"
)

/*

1) test if external balance of ETH is getting updated first time
2) test if external ETH is greater than already existing eth
*/
func TestInit(t *testing.T) {
	var (
		accountCache *cache.Cache
		eBalances    map[string]statedb.EBalance
	)
	accountCache = cache.New()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(0)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	es := &EthSyncer{Account: account, Cache: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(1)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)
	assert.Equal(t, cachedAcc.(cache.AccountCache).Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Balance should be updated with external balance")

}
func TestExternalETHisGreater(t *testing.T) {
	var (
		accountCache *cache.Cache
		eBalances    map[string]statedb.EBalance
	)
	accountCache = cache.New()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	es := &EthSyncer{Account: account, Cache: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(3)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.(cache.AccountCache).Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.(cache.AccountCache).LastExtBalance["ETH"], big.NewInt(3), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).CurrentExtBalance["ETH"], big.NewInt(3), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).IsFirstEntry["ETH"], true, "CurrentExtBalance ahould be updated")

	es.ExtBalance = big.NewInt(20)
	es.Update()
	assert.Equal(t, cachedAcc.(cache.AccountCache).Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.(cache.AccountCache).IsFirstEntry["ETH"], false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).IsNewAmountUpdate["ETH"], true, "CurrentExtBalance ahould be updated")

	assert.Equal(t, cachedAcc.(cache.AccountCache).LastExtBalance["ETH"], big.NewInt(20), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).CurrentExtBalance["ETH"], big.NewInt(20), "CurrentExtBalance ahould be updated")
}

func TestExternalETHisLesser(t *testing.T) {
	var (
		accountCache *cache.Cache
		eBalances    map[string]statedb.EBalance
	)
	accountCache = cache.New()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	es := &EthSyncer{Account: account, Cache: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.(cache.AccountCache).Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "should be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).LastExtBalance["ETH"], big.NewInt(10), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).CurrentExtBalance["ETH"], big.NewInt(10), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).IsFirstEntry["ETH"], true, "CurrentExtBalance should be updated")

	es.ExtBalance = big.NewInt(1)
	es.Update()
	assert.Equal(t, cachedAcc.(cache.AccountCache).IsFirstEntry["ETH"], false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).IsNewAmountUpdate["ETH"], true, "CurrentExtBalance ahould be updated")

	assert.Equal(t, cachedAcc.(cache.AccountCache).LastExtBalance["ETH"], big.NewInt(1), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).CurrentExtBalance["ETH"], big.NewInt(1), "CurrentExtBalance ahould be updated")
}

// Test if cache exists but Account cache dont have eth asset
func Test(t *testing.T) {
	var (
		accountCache *cache.Cache
		eBalances    map[string]statedb.EBalance
	)
	accountCache = cache.New()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress"

	accountCache.Set(account.Address, cache.AccountCache{})

	es := &EthSyncer{Account: account, Cache: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)
	//ac := cachedAcc.(cache.AccountCache).UpdateLastExtBalanceByKey("testEthAddress", big.NewInt(10))

	//assert.Equal(t, cachedAcc.(cache.AccountCache).Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "should be updated")
	assert.Equal(t, cachedAcc.(cache.AccountCache).LastExtBalance["ETH"], big.NewInt(10), "should be updated")

}
