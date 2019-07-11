package sync

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"

	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

// HBTCSyncer syncs all HBTC external accounts
// HBCT account is the first ETH account of user
type HBTCSyncer struct {
	LastExtBalance map[string]*big.Int
	ExtBalance     map[string]*big.Int
	RPC            string
	Account        statedb.Account
	Storage        external.BalanceStorage
}

func newHBTCSyncer() *HBTCSyncer {
	e := &HBTCSyncer{}
	e.ExtBalance = make(map[string]*big.Int)
	e.LastExtBalance = make(map[string]*big.Int)

	return e
}

// GetExtBalance ...
func (hs *HBTCSyncer) GetExtBalance() error {
	// If ETH account exists
	ethAccount, ok := hs.Account.EBalances["ETH"]
	if !ok {
		return errors.New("ETH account does not exists")
	}

	hbtcAccount, ok := ethAccount[hs.Account.FirstExternalAddress["ETH"]]
	if !ok {
		return errors.New("HBTC account does not exists")
	}

	resp, err := http.Get(fmt.Sprintf("%s/%s", hs.RPC, hbtcAccount.Address))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	balance, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return err
	}

	hs.ExtBalance[hbtcAccount.Address] = big.NewInt(balance)

	return nil
}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (hs *HBTCSyncer) Update() {
	assetSymbol := "HBTC"
	if hs.Account.EBalances[assetSymbol] == nil {
		log.Println("No HBTC account available, skip")
		return
	}

	// HBTC account is first ETH account of user.
	ethAccount := hs.Account.EBalances[assetSymbol][hs.Account.FirstExternalAddress["ETH"]]
	hBTCBalance := *big.NewInt(int64(0))
	storageKey := assetSymbol + "-" + ethAccount.Address
	if last, ok := hs.Storage.Get(hs.Account.Address); ok {
		// last-balance < External-ETH
		// Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
		if lastExtBalance, ok := last.LastExtBalance[storageKey]; ok && lastExtBalance != nil {
			if lastExtBalance.Cmp(hs.ExtBalance[ethAccount.Address]) < 0 {
				log.Printf("lastExtBalance.Cmp(hs.ExtBalance[%s])", ethAccount.Address)

				hBTCBalance.Sub(hs.ExtBalance[ethAccount.Address], lastExtBalance)

				ethAccount.Balance += hBTCBalance.Uint64()
				hs.Account.EBalances[assetSymbol][ethAccount.Address] = ethAccount

				last = last.UpdateLastExtBalanceByKey(storageKey, hs.ExtBalance[ethAccount.Address])
				last = last.UpdateCurrentExtBalanceByKey(storageKey, hs.ExtBalance[ethAccount.Address])
				last = last.UpdateIsFirstEntryByKey(storageKey, false)
				last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
				last = last.UpdateAccount(hs.Account)
				hs.Storage.Set(hs.Account.Address, last)
				log.Printf("New account balance after hbtc balance credit: %v\n", last)
			}

			// last-balance < External-ETH
			// Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
			if lastExtBalance.Cmp(hs.ExtBalance[ethAccount.Address]) > 0 {
				log.Println("lastExtBalance.Cmp(hs.ExtBalance) ============")

				hBTCBalance.Sub(lastExtBalance, hs.ExtBalance[ethAccount.Address])

				ethAccount.Balance -= hBTCBalance.Uint64()
				hs.Account.EBalances[assetSymbol][ethAccount.Address] = ethAccount

				last = last.UpdateLastExtBalanceByKey(storageKey, hs.ExtBalance[ethAccount.Address])
				last = last.UpdateCurrentExtBalanceByKey(storageKey, hs.ExtBalance[ethAccount.Address])
				last = last.UpdateIsFirstEntryByKey(storageKey, false)
				last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
				last = last.UpdateAccount(hs.Account)
				hs.Storage.Set(hs.Account.Address, last)
				log.Printf("New account balance after hbtc balance debit: %v\n", last)
			}
			return
		}

		log.Printf("Initialise external balance in cache: %v\n", last)
		if hs.ExtBalance[ethAccount.Address] == nil {
			hs.ExtBalance[ethAccount.Address] = big.NewInt(0)
		}
		last = last.UpdateLastExtBalanceByKey(storageKey, hs.ExtBalance[ethAccount.Address])
		last = last.UpdateCurrentExtBalanceByKey(storageKey, hs.ExtBalance[ethAccount.Address])
		last = last.UpdateIsFirstEntryByKey(storageKey, true)
		last = last.UpdateIsNewAmountUpdateByKey(storageKey, false)
		ethAccount.UpdateBalance(hs.ExtBalance[ethAccount.Address].Uint64())
		hs.Account.EBalances[assetSymbol][ethAccount.Address] = ethAccount
		last = last.UpdateAccount(hs.Account)
		hs.Storage.Set(hs.Account.Address, last)
		return
	}

	log.Printf("Initialise account in cache.")
	balance := hs.ExtBalance[ethAccount.Address]
	lastbalances := make(map[string]*big.Int)
	lastbalances[storageKey] = hs.ExtBalance[ethAccount.Address]

	currentbalances := make(map[string]*big.Int)
	currentbalances[storageKey] = hs.ExtBalance[ethAccount.Address]
	if balance == nil {
		lastbalances[storageKey] = big.NewInt(0)
		currentbalances[storageKey] = big.NewInt(0)
	}
	isFirstEntry := make(map[string]bool)
	isFirstEntry[storageKey] = true
	isNewAmountUpdate := make(map[string]bool)
	isNewAmountUpdate[storageKey] = false
	if balance != nil {
		ethAccount.UpdateBalance(balance.Uint64())
	}

	hs.Account.EBalances[assetSymbol][ethAccount.Address] = ethAccount
	val := external.AccountCache{
		Account:           hs.Account,
		LastExtBalance:    lastbalances,
		CurrentExtBalance: currentbalances,
		IsFirstEntry:      isFirstEntry,
		IsNewAmountUpdate: isNewAmountUpdate,
	}
	hs.Storage.Set(hs.Account.Address, val)
}
