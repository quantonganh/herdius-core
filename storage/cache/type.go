package cache

import (
	"math/big"

	"github.com/herdius/herdius-core/storage/state/statedb"
)

// AccountCache holds the balance detail of an Account
// that we need to use while updating balances of external assets
type AccountCache struct {
	Account           statedb.Account
	LastExtBalance    map[string]*big.Int
	CurrentExtBalance map[string]*big.Int
	IsFirstEntry      map[string]bool
}
