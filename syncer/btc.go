package sync

import (
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"

	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

type BTCSyncer struct {
	LastExtBalance *big.Int
	ExtBalance     *big.Int
	Account        statedb.Account
	ExBal          external.BalanceStorage
	RPC            string
}

func (es *BTCSyncer) GetExtBalance() {
	var url string

	btcAccount, ok := es.Account.EBalances["BTC"]
	if !ok {
		return
	}

	apiKey := os.Getenv("BLOCKCHAIN_INFO_KEY")
	if len(apiKey) > 0 {
		url = es.RPC + "/addressbalance/" + btcAccount.Address + "?api_code=" + apiKey
	} else {
		url = es.RPC + "/addressbalance/" + btcAccount.Address
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error connecting Blockchain info ", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		balance := new(big.Int)
		balance.SetString(bodyString, 10)
		es.ExtBalance = balance
	}

}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (es *BTCSyncer) Update() {
	assetSymbol := "BTC"
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
					log.Println("Last balance is less that external for asset", assetSymbol)
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
					log.Println("Last balance is greater that external for asset", assetSymbol)

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
				log.Printf("Initialise external balance in cache: %v\n", last)

				last = last.UpdateLastExtBalanceByKey(assetSymbol, es.ExtBalance)
				last = last.UpdateCurrentExtBalanceByKey(assetSymbol, es.ExtBalance)
				last = last.UpdateIsFirstEntryByKey(assetSymbol, true)
				last = last.UpdateIsNewAmountUpdateByKey(assetSymbol, false)
				value.UpdateBalance(es.ExtBalance.Uint64())
				es.Account.EBalances[assetSymbol] = value
				last = last.UpdateAccount(es.Account)

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
			value.UpdateBalance(es.ExtBalance.Uint64())
			es.Account.EBalances[assetSymbol] = value
			val := external.AccountCache{
				Account: es.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
			}
			es.ExBal.Set(es.Account.Address, val)
		}

	}
}
