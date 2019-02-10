package statedb

import (
	cmn "github.com/herdius/herdius-core/libs/common"
)

// Account : Account Detail
type Account struct {
	Nonce       uint64
	Address     string
	AddressHash cmn.HexBytes
	Balance     []byte
}
