package sync

import (
	"context"
	"fmt"
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
	value, ok := es.Account.EBalances[assetSymbol]
	if ok {
		herEthBalance := *big.NewInt(int64(0))
		last, ok := es.Cache.Get(es.Account.Address)
		if ok {
			//last-balance < External-ETH
			//Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
			fmt.Printf("Address: %v\n", es.Account.Address)
			fmt.Printf("es.ExtBalance : %v\n", es.ExtBalance)
			fmt.Printf("last.(cache.AccountCache) : %v\n", last.(cache.AccountCache).LastExtBalance)
			lastExtBalance, ok := last.(cache.AccountCache).LastExtBalance[assetSymbol]
			if ok {
				if lastExtBalance.Cmp(es.ExtBalance) < 0 {
					herEthBalance.Sub(es.ExtBalance, lastExtBalance)
					value.Balance += herEthBalance.Uint64()
					es.Account.EBalances[assetSymbol] = value

					last = last.(cache.AccountCache).UpdateLastExtBalanceByKey(assetSymbol, es.ExtBalance)
					last = last.(cache.AccountCache).UpdateCurrentExtBalanceByKey(assetSymbol, es.ExtBalance)
					last = last.(cache.AccountCache).UpdateIsFirstEntryByKey(assetSymbol, false)
					last = last.(cache.AccountCache).UpdateIsNewAmountUpdateByKey(assetSymbol, true)

					log.Printf("New account balance after external balance credit: %v\n", last)
					es.Cache.Set(es.Account.Address, last)
					return

				}

				//last-balance < External-ETH
				//Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
				if lastExtBalance.Cmp(es.ExtBalance) > 0 {
					herEthBalance.Sub(lastExtBalance, es.ExtBalance)
					value.Balance -= herEthBalance.Uint64()
					last = last.(cache.AccountCache).UpdateLastExtBalanceByKey(assetSymbol, es.ExtBalance)
					last = last.(cache.AccountCache).UpdateCurrentExtBalanceByKey(assetSymbol, es.ExtBalance)
					last = last.(cache.AccountCache).UpdateIsFirstEntryByKey(assetSymbol, false)
					last = last.(cache.AccountCache).UpdateIsNewAmountUpdateByKey(assetSymbol, true)
					last = last.(cache.AccountCache).UpdateAccount(es.Account)

					log.Printf("New account balance after external balance debit: %v\n", last)
					es.Cache.Set(es.Account.Address, last)
					return
				}
			} else {
				last = last.(cache.AccountCache).UpdateLastExtBalanceByKey(assetSymbol, es.ExtBalance)
				last = last.(cache.AccountCache).UpdateCurrentExtBalanceByKey(assetSymbol, es.ExtBalance)
				last = last.(cache.AccountCache).UpdateIsFirstEntryByKey(assetSymbol, true)
				last = last.(cache.AccountCache).UpdateIsNewAmountUpdateByKey(assetSymbol, false)
				value.UpdateBalance(es.ExtBalance.Uint64())
				es.Account.EBalances[assetSymbol] = value
				last = last.(cache.AccountCache).UpdateAccount(es.Account)

				es.Cache.Set(es.Account.Address, last)
			}

		} else {

			lastbalances := make(map[string]*big.Int)
			lastbalances[assetSymbol] = es.ExtBalance
			currentbalances := make(map[string]*big.Int)
			currentbalances[assetSymbol] = es.ExtBalance
			isFirstEntry := make(map[string]bool)
			isFirstEntry[assetSymbol] = true
			isNewAmountUpdate := make(map[string]bool)
			isNewAmountUpdate[assetSymbol] = false

			log.Println("New address will be updated with external balance")
			value.UpdateBalance(es.ExtBalance.Uint64())
			es.Account.EBalances[assetSymbol] = value
			val := cache.AccountCache{
				Account: es.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
			}
			es.Cache.Set(es.Account.Address, val)
		}

	}
}
