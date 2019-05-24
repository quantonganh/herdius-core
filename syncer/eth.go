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

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (es *EthSyncer) Update() {
	assetSymbol := "ETH"
	value, ok := es.Account.EBalances["ETH"]
	if ok {
		herEthBalance := *big.NewInt(int64(0))
		last, ok := es.Cache.Get(es.Account.Address)

		if ok {
			//last-balance < External-ETH
			//Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)

			lastAssetBalance, ok := last.(cache.AccountCache).LastExtBalance[assetSymbol]
			if ok {
				if lastAssetBalance.Cmp(es.ExtBalance) < 0 {
					herEthBalance.Sub(es.ExtBalance, lastAssetBalance)
					value.Balance += herEthBalance.Uint64()
					es.Account.EBalances["ETH"] = value
					last.(cache.AccountCache).LastExtBalance[assetSymbol] = es.ExtBalance
					last.(cache.AccountCache).CurrentExtBalance[assetSymbol] = es.ExtBalance
					last.(cache.AccountCache).IsFirstEntry[assetSymbol] = false

					val := cache.AccountCache{
						Account: es.Account, LastExtBalance: last.(cache.AccountCache).LastExtBalance, CurrentExtBalance: last.(cache.AccountCache).CurrentExtBalance, IsFirstEntry: last.(cache.AccountCache).IsFirstEntry,
					}
					es.Cache.Set(es.Account.Address, val)
					return

				}

				//last-balance < External-ETH
				//Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
				lastAssetBalance, _ = last.(cache.AccountCache).LastExtBalance[assetSymbol]

				if lastAssetBalance.Cmp(es.ExtBalance) > 0 {
					herEthBalance.Sub(lastAssetBalance, es.ExtBalance)
					value.Balance -= herEthBalance.Uint64()
					es.Account.EBalances["ETH"] = value
					last.(cache.AccountCache).LastExtBalance[assetSymbol] = es.ExtBalance
					last.(cache.AccountCache).CurrentExtBalance[assetSymbol] = es.ExtBalance
					last.(cache.AccountCache).IsFirstEntry[assetSymbol] = false

					val := cache.AccountCache{
						Account: es.Account, LastExtBalance: last.(cache.AccountCache).LastExtBalance, CurrentExtBalance: last.(cache.AccountCache).CurrentExtBalance, IsFirstEntry: last.(cache.AccountCache).IsFirstEntry,
					}
					es.Cache.Set(es.Account.Address, val)
					return
				}

			} else {
				last.(cache.AccountCache).LastExtBalance[assetSymbol] = es.ExtBalance
				last.(cache.AccountCache).CurrentExtBalance[assetSymbol] = es.ExtBalance
				last.(cache.AccountCache).IsFirstEntry[assetSymbol] = true
				value.UpdateBalance(es.ExtBalance.Uint64())
				es.Account.EBalances["ETH"] = value
				val := cache.AccountCache{
					Account: es.Account, LastExtBalance: last.(cache.AccountCache).LastExtBalance, CurrentExtBalance: last.(cache.AccountCache).CurrentExtBalance, IsFirstEntry: last.(cache.AccountCache).IsFirstEntry,
				}
				es.Cache.Set(es.Account.Address, val)
			}

		} else {
			lastbalances := make(map[string]*big.Int)
			lastbalances["ETH"] = es.ExtBalance
			currentbalances := make(map[string]*big.Int)
			currentbalances["ETH"] = es.ExtBalance
			isFirstEntry := make(map[string]bool)
			isFirstEntry["ETH"] = true

			value.UpdateBalance(es.ExtBalance.Uint64())
			es.Account.EBalances["ETH"] = value
			val := cache.AccountCache{
				Account: es.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry,
			}
			es.Cache.Set(es.Account.Address, val)

		}

	}

}
