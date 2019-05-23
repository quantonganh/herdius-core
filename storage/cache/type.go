package cache

import (
	"math/big"

	"github.com/herdius/herdius-core/storage/state/statedb"
)

// AccountCache holds the balance detail of an Account
// that we need to be used while updating the external asset
// balances
type AccountCache struct {
	Account           statedb.Account
	LastExtBalance    *big.Int
	CurrentExtBalance *big.Int
	IsFirstEntry      bool
}
