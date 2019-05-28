package sync

import (
	"math/big"
	"testing"

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
	accountCache = external.NewTest()
	defer accountCache.CloseTest()

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
	accountCache = external.NewTest()
	defer accountCache.CloseTest()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress5"

	es := &EthSyncer{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(3)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(3), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(3), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], true, "CurrentExtBalance ahould be updated")

	es.ExtBalance = big.NewInt(10)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["ETH"], true, "CurrentExtBalance ahould be updated")

	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "Total fetched transactions should be 20")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(10), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(10), "CurrentExtBalance ahould be updated")
}

func TestHERExternalETHisLesser(t *testing.T) {
	var (
		accountCache external.BalanceStorage
		eBalances    map[string]statedb.EBalance
	)
	accountCache = external.NewTest()
	defer accountCache.CloseTest()

	eBalances = make(map[string]statedb.EBalance)
	eBalances["ETH"] = statedb.EBalance{Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testEthAddress00"

	es := &EthSyncer{Account: account, ExBal: accountCache}
	// Set external balance coming from infura
	es.ExtBalance = big.NewInt(10)
	es.Update()
	cachedAcc, _ := accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.Account.EBalances["ETH"].Balance, es.ExtBalance.Uint64(), "should be updated")
	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(10), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(10), "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], true, "CurrentExtBalance ahould be updated")

	es.ExtBalance = big.NewInt(1)
	es.Update()
	cachedAcc, _ = accountCache.Get(account.Address)

	assert.Equal(t, cachedAcc.IsFirstEntry["ETH"], false, "CurrentExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.IsNewAmountUpdate["ETH"], true, "CurrentExtBalance ahould be updated")

	assert.Equal(t, cachedAcc.LastExtBalance["ETH"], big.NewInt(1), "LastExtBalance ahould be updated")
	assert.Equal(t, cachedAcc.CurrentExtBalance["ETH"], big.NewInt(1), "CurrentExtBalance ahould be updated")
}
