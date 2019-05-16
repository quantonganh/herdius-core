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
	EBalances   map[string]EBalance
}

// EBalance is external balance model
type EBalance struct {
	Address         string
	Balance         uint64
	LastBlockHeight uint64
	Nonce           uint64
}
