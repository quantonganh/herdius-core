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
	LastExtBalance *big.Int
	ExtBalance     *big.Int
	Account        statedb.Account
	Cache          *cache.Cache
	RPC            string
}

func (es *EthSyncer) GetExtBalance() {
	client, err := ethclient.Dial(es.RPC)
	if err != nil {
		log.Println("Error connecting ETH RPC", err)
	}
	// If ETH account exists
	ethAccount, ok := es.Account.EBalances["ETH"]
	if !ok {

		return
	}

	account := common.HexToAddress(ethAccount.Address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Println("Error getting ETH Balance from RPC", err)

	}
	es.ExtBalance = balance
}

func (es *EthSyncer) Update() {
	value, ok := es.Account.EBalances["ETH"]
	if ok {
		herEthBalance := *big.NewInt(int64(0))
		last, ok := es.Cache.Get(es.Account.Address)
		if ok {
			//last-balance < External-ETH
			if last.(cache.AccountCache).LastExtBalance.Cmp(es.ExtBalance) < 0 {
				//herEth = exteth - lastEth
				herEthBalance.Sub(es.ExtBalance, last.(cache.AccountCache).LastExtBalance)
				herEthBalance.Add(&herEthBalance, es.ExtBalance)
				value.UpdateBalance(herEthBalance.Uint64())
				es.Account.EBalances["ETH"] = value
				val := cache.AccountCache{Account: es.Account, LastExtBalance: es.ExtBalance}
				es.Cache.Set(es.Account.Address, val)
				return

			}
			if last.(cache.AccountCache).LastExtBalance.Cmp(es.ExtBalance) == 0 {
				value.UpdateBalance(es.ExtBalance.Uint64())
				es.Account.EBalances["ETH"] = value
				val := cache.AccountCache{Account: es.Account, LastExtBalance: es.ExtBalance}
				es.Cache.Set(es.Account.Address, val)
				return
			}
		} else {
			value.UpdateBalance(es.ExtBalance.Uint64())
			es.Account.EBalances["ETH"] = value
			val := cache.AccountCache{Account: es.Account, LastExtBalance: es.ExtBalance}
			es.Cache.Set(es.Account.Address, val)

		}

	}

}