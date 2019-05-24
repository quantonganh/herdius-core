package sync

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/herdius/herdius-core/storage/cache"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/herdius/herdius-core/syncer/contract"
)

type HERToken struct {
	LastExtBalance       *big.Int
	ExtBalance           *big.Int
	Account              statedb.Account
	Cache                *cache.Cache
	TokenContractAddress string
	TokenSymbol          string
	RPC                  string
}

//GetExtBalance Gets Asset balance from main chain
func (es *HERToken) GetExtBalance() {
	client, err := ethclient.Dial(es.RPC)
	if err != nil {
		log.Println("Error connecting ETH RPC", err)
	}
	tokenAddress := common.HexToAddress(es.TokenContractAddress)

	instance, err := contract.NewToken(tokenAddress, client)
	if err != nil {
		log.Fatal(err)
	}
	address := common.HexToAddress(es.Account.Erc20Address)
	bal, err := instance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatal(err)
	}

	es.ExtBalance = bal
}

//Update Updates balance of asset in cache
func (es *HERToken) Update() {
	herBalance := *big.NewInt(int64(0))
	last, ok := es.Cache.Get(es.Account.Address)
	if ok {
		//last-balance < External-ETH
		//Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
		fmt.Printf("Address: %v\n", es.Account.Address)
		fmt.Printf("es.ExtBalance : %v\n", es.ExtBalance)
		fmt.Printf("last.(cache.AccountCache) : %v\n", last.(cache.AccountCache).LastExtBalance)
		lastExtHERBalance := last.(cache.AccountCache).LastExtHERBalance
		if lastExtHERBalance != nil {
			if lastExtHERBalance.Cmp(es.ExtBalance) < 0 {
				herBalance.Sub(es.ExtBalance, lastExtHERBalance)
				es.Account.Balance += herBalance.Uint64()
				// val := cache.AccountCache{
				// 	Account:              es.Account,
				// 	LastExtBalance:       last.(cache.AccountCache).LastExtBalance,
				// 	CurrentExtBalance:    last.(cache.AccountCache).CurrentExtBalance,
				// 	IsFirstEntry:         last.(cache.AccountCache).IsFirstEntry,
				// 	IsNewAmountUpdate:    last.(cache.AccountCache).IsNewAmountUpdate,
				// 	LastExtHERBalance:    es.ExtBalance,
				// 	CurrentExtHERBalance: es.ExtBalance,
				// 	IsFirstHEREntry:      false,
				// 	IsNewHERAmountUpdate: true,
				// }

				last = last.(cache.AccountCache).UpdateAccount(es.Account)
				last = last.(cache.AccountCache).UpdateLastExtHERBalance(es.ExtBalance)
				last = last.(cache.AccountCache).UpdateCurrentExtHERBalance(es.ExtBalance)
				last = last.(cache.AccountCache).UpdateIsNewHERAmountUpdate(true)
				last = last.(cache.AccountCache).UpdateIsFirstHER(false)

				log.Printf("New account balance after external balance credit: %v\n", last)
				es.Cache.Set(es.Account.Address, last)
				return

			}

			//last-balance < External-ETH
			//Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
			if lastExtHERBalance.Cmp(es.ExtBalance) > 0 {
				herBalance.Sub(lastExtHERBalance, es.ExtBalance)
				es.Account.Balance -= herBalance.Uint64()
				// val := cache.AccountCache{
				// 	Account:              es.Account,
				// 	LastExtBalance:       last.(cache.AccountCache).LastExtBalance,
				// 	CurrentExtBalance:    last.(cache.AccountCache).CurrentExtBalance,
				// 	IsFirstEntry:         last.(cache.AccountCache).IsFirstEntry,
				// 	IsNewAmountUpdate:    last.(cache.AccountCache).IsNewAmountUpdate,
				// 	LastExtHERBalance:    es.ExtBalance,
				// 	CurrentExtHERBalance: es.ExtBalance,
				// 	IsFirstHEREntry:      false,
				// 	IsNewHERAmountUpdate: true,
				// }

				last = last.(cache.AccountCache).UpdateAccount(es.Account)
				last = last.(cache.AccountCache).UpdateLastExtHERBalance(es.ExtBalance)
				last = last.(cache.AccountCache).UpdateCurrentExtHERBalance(es.ExtBalance)
				last = last.(cache.AccountCache).UpdateIsNewHERAmountUpdate(true)
				last = last.(cache.AccountCache).UpdateIsFirstHER(false)

				log.Printf("New account balance after external balance debit: %v\n", last)
				es.Cache.Set(es.Account.Address, last)
				return
			}
		} else {

			es.Account.Balance = es.ExtBalance.Uint64()

			last = last.(cache.AccountCache).UpdateAccount(es.Account)
			last = last.(cache.AccountCache).UpdateIsFirstHER(true)
			last = last.(cache.AccountCache).UpdateLastExtHERBalance(es.ExtBalance)
			last = last.(cache.AccountCache).UpdateCurrentExtHERBalance(es.ExtBalance)

			es.Cache.Set(es.Account.Address, last)

		}

	}

}
