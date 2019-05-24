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
