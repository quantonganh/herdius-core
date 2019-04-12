package statedb

import (
	cmn "github.com/herdius/herdius-core/libs/common"
)

// Account : Account Detail
type Account struct {
	Nonce       uint64
	Address     string
	PublicKey   string
	StateRoot   string
	AddressHash cmn.HexBytes
	Balance     uint64
	Balances    map[string]uint64 // Balances will store balances of assets e.g. [BTC]=10 or [HER]=1000
}
