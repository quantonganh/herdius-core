package cache

import (
	"math/big"

	"github.com/herdius/herdius-core/storage/state/statedb"
)

type AccountCache struct {
	Account        statedb.Account
	LastExtBalance *big.Int
}
