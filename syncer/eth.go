package sync

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

type EthSyncer struct {
	LastExtBalance *big.Int
	ExtBalance     *big.Int
	Account        statedb.Account
	ExBal          external.BalanceStorage
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
		last, ok := es.ExBal.Get(es.Account.Address)
		if ok {
			//last-balance < External-ETH
			//Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
			lastExtBalance, ok := last.LastExtBalance[assetSymbol]
			if ok {
				if lastExtBalance.Cmp(es.ExtBalance) < 0 {
					log.Println("lastExtBalance.Cmp(es.ExtBalance)")
					herEthBalance.Sub(es.ExtBalance, lastExtBalance)
					value.Balance += herEthBalance.Uint64()
					es.Account.EBalances[assetSymbol] = value

					last = last.UpdateLastExtBalanceByKey(assetSymbol, es.ExtBalance)
					last = last.UpdateCurrentExtBalanceByKey(assetSymbol, es.ExtBalance)
					last = last.UpdateIsFirstEntryByKey(assetSymbol, false)
					last = last.UpdateIsNewAmountUpdateByKey(assetSymbol, true)
					last = last.UpdateAccount(es.Account)

					log.Printf("New account balance after external balance credit: %v\n", last)
					es.ExBal.Set(es.Account.Address, last)
					return

				}

				//last-balance < External-ETH
				//Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
				if lastExtBalance.Cmp(es.ExtBalance) > 0 {
					log.Println("lastExtBalance.Cmp(es.ExtBalance) ============")

					herEthBalance.Sub(lastExtBalance, es.ExtBalance)
					value.Balance -= herEthBalance.Uint64()
					last = last.UpdateLastExtBalanceByKey(assetSymbol, es.ExtBalance)
					last = last.UpdateCurrentExtBalanceByKey(assetSymbol, es.ExtBalance)
					last = last.UpdateIsFirstEntryByKey(assetSymbol, false)
					last = last.UpdateIsNewAmountUpdateByKey(assetSymbol, true)
					es.Account.EBalances[assetSymbol] = value
					last = last.UpdateAccount(es.Account)

					log.Printf("New account balance after external balance debit: %v\n", last)
					es.ExBal.Set(es.Account.Address, last)
					return
				}
			} else {
				log.Println("--- ============")

				last = last.UpdateLastExtBalanceByKey(assetSymbol, es.ExtBalance)
				last = last.UpdateCurrentExtBalanceByKey(assetSymbol, es.ExtBalance)
				last = last.UpdateIsFirstEntryByKey(assetSymbol, true)
				last = last.UpdateIsNewAmountUpdateByKey(assetSymbol, false)
				value.UpdateBalance(es.ExtBalance.Uint64())
				es.Account.EBalances[assetSymbol] = value
				last = last.UpdateAccount(es.Account)

				log.Println("--- ============", last)

				es.ExBal.Set(es.Account.Address, last)
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
			val := external.AccountCache{
				Account: es.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
			}
			es.ExBal.Set(es.Account.Address, val)
		}

	}
}
