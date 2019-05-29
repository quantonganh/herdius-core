package sync

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/herdius/herdius-core/syncer/contract"
)

type HERToken struct {
	LastExtBalance       *big.Int
	ExtBalance           *big.Int
	Account              statedb.Account
	ExBal                external.BalanceStorage
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
	last, ok := es.ExBal.Get(es.Account.Address)
	if ok {
		//last-balance < External-ETH
		//Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
		lastExtHERBalance := last.LastExtHERBalance
		if lastExtHERBalance != nil {
			if lastExtHERBalance.Cmp(es.ExtBalance) < 0 {
				herBalance.Sub(es.ExtBalance, lastExtHERBalance)
				es.Account.Balance += herBalance.Uint64()

				last = last.UpdateAccount(es.Account)
				last = last.UpdateLastExtHERBalance(es.ExtBalance)
				last = last.UpdateCurrentExtHERBalance(es.ExtBalance)
				last = last.UpdateIsNewHERAmountUpdate(true)
				last = last.UpdateIsFirstHER(false)

				log.Printf("New account balance after external balance credit: %v\n", last)
				es.ExBal.Set(es.Account.Address, last)
				return

			}

			//last-balance < External-ETH
			//Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
			if lastExtHERBalance.Cmp(es.ExtBalance) > 0 {
				herBalance.Sub(lastExtHERBalance, es.ExtBalance)
				es.Account.Balance -= herBalance.Uint64()
				last = last.UpdateAccount(es.Account)
				last = last.UpdateLastExtHERBalance(es.ExtBalance)
				last = last.UpdateCurrentExtHERBalance(es.ExtBalance)
				last = last.UpdateIsNewHERAmountUpdate(true)
				last = last.UpdateIsFirstHER(false)

				log.Printf("New account balance after external balance debit: %v\n", last)
				es.ExBal.Set(es.Account.Address, last)
				return
			}
		} else {

			es.Account.Balance = es.ExtBalance.Uint64()

			last = last.UpdateAccount(es.Account)
			last = last.UpdateIsFirstHER(true)
			last = last.UpdateLastExtHERBalance(es.ExtBalance)
			last = last.UpdateCurrentExtHERBalance(es.ExtBalance)

			es.ExBal.Set(es.Account.Address, last)

		}

	}

}
