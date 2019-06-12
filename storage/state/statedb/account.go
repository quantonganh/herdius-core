package statedb

import (
	cmn "github.com/herdius/herdius-core/libs/common"
)

// Account : Account Detail
type Account struct {
	Nonce           uint64
	Address         string
	PublicKey       string
	StateRoot       string
	AddressHash     cmn.HexBytes
	Balance         uint64
	Erc20Address    string
	LastBlockHeight uint64
	ExternalNonce   uint64
	EBalances       map[string]map[string]EBalance
}

// EBalance is external balance model
type EBalance struct {
	Address         string
	Balance         uint64
	LastBlockHeight uint64
	Nonce           uint64
}

// UpdateBalance sets EBalance's balance.
func (eb *EBalance) UpdateBalance(b uint64) {
	eb.Balance = b
}

// UpdateBlockHeight sets EBalance's last block height.
func (eb *EBalance) UpdateBlockHeight(h uint64) {
	eb.LastBlockHeight = h
}

// UpdateNonce sets EBalance's nonce.
func (eb *EBalance) UpdateNonce(n uint64) {
	eb.Nonce = n
}
