package cache

import (
	"math/big"

	"github.com/herdius/herdius-core/storage/state/statedb"
)

// AccountCache holds the balance detail of an Account
// that we need to use while updating balances of external assets
type AccountCache struct {
	Account              statedb.Account
	LastExtHERBalance    *big.Int
	CurrentExtHERBalance *big.Int
	IsFirstHEREntry      bool
	IsNewHERAmountUpdate bool
	LastExtBalance       map[string]*big.Int
	CurrentExtBalance    map[string]*big.Int
	IsFirstEntry         map[string]bool
	IsNewAmountUpdate    map[string]bool
}

func (ac AccountCache) UpdateAccount(account statedb.Account) {
	ac.Account = account
}
func (ac AccountCache) UpdateLastExtHERBalance(lehb *big.Int) {
	ac.LastExtHERBalance = lehb
}

func (ac AccountCache) UpdateCurrentExtHERBalance(lchb *big.Int) {
	ac.CurrentExtHERBalance = lchb
}

func (ac AccountCache) UpdateIsFirstHER(isFirst bool) {
	ac.IsFirstHEREntry = isFirst
}

func (ac AccountCache) UpdateIsNewHERAmountUpdate(isnew bool) {
	ac.IsNewHERAmountUpdate = isnew
}

func (ac AccountCache) UpdateLastExtBalance(exbalance map[string]*big.Int) {
	ac.LastExtBalance = exbalance
}

func (ac AccountCache) UpdateCurrentExtBalance(cbalance map[string]*big.Int) {
	ac.CurrentExtBalance = cbalance
}
func (ac AccountCache) UpdateIsFirstEntry(isfirst map[string]bool) {
	ac.IsFirstEntry = isfirst
}
func (ac AccountCache) UpdateIsNewAmountUpdate(isNew map[string]bool) {
	ac.IsFirstEntry = isNew
}
