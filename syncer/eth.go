package sync

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/herdius/herdius-core/storage/cache"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

type EthSyncer struct {
	ExtBalance *big.Int
	Account    statedb.Account
	Cache      *cache.Cache
}

func (es *EthSyncer) GetExtBalance() {
	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		log.Fatal(err)
	}

	// If ETH account exists
	ethAccount, ok := es.Account.EBalances["ETH"]
	if !ok {
		return
	}

	account := common.HexToAddress(ethAccount.Address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	es.ExtBalance = balance
}

func (es *EthSyncer) Update() {
	value, ok := es.Account.EBalances["ETH"]
	if ok {
		value.UpdateBalance(es.ExtBalance.Uint64())
		es.Account.EBalances["ETH"] = value
		es.Cache.Set(es.Account.Address, es.Account)

	}

}
