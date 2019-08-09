package sync

import (
	"errors"
	"log"
	"math/big"
	"time"

	blockcypher "github.com/blockcypher/gobcy"

	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

// BTCSyncer syncs all external BTC accounts.
type BTCTestNetSyncer struct {
	LastExtBalance map[string]*big.Int
	ExtBalance     map[string]*big.Int
	BlockHeight    map[string]*big.Int
	Nonce          map[string]uint64
	RPC            string
	Account        statedb.Account
	Storage        external.BalanceStorage
	addressError   map[string]bool
}

func newBTCTestNetSyncer() *BTCTestNetSyncer {
	b := &BTCTestNetSyncer{}
	b.ExtBalance = make(map[string]*big.Int)
	b.LastExtBalance = make(map[string]*big.Int)
	b.BlockHeight = make(map[string]*big.Int)
	b.Nonce = make(map[string]uint64)
	b.addressError = make(map[string]bool)

	return b
}

// GetExtBalance ...
func (btc *BTCTestNetSyncer) GetExtBalance() error {

	btcAccount, ok := btc.Account.EBalances["BTC"]
	if !ok {
		return errors.New("BTC account does not exists")
	}
	btcCypher := blockcypher.API{Token: "490bb2949a2542fcb6f74f4efdba70dd", Coin: "btc", Chain: "test3"}
	// Blockcypher should send 3 requests every 15 seconds since we are using
	// free service
	count := 1
	for _, ba := range btcAccount {
		if count == 3 {
			time.Sleep(15 * time.Second)
			count = 1
		}
		addr, err := btcCypher.GetAddrFull(ba.Address, nil)
		if err != nil {
			log.Println("Error getting BTC address", err)
			btc.addressError[ba.Address] = true
			continue
		}
		if len(addr.TXs) > 0 {
			btc.BlockHeight[ba.Address] = big.NewInt(int64(addr.TXs[0].BlockHeight))
			btc.Nonce[ba.Address] = uint64(len(addr.TXs))
			btc.ExtBalance[ba.Address] = big.NewInt(int64(addr.Balance))
			btc.addressError[ba.Address] = false
		}
		count++

	}
	return nil

}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (btc *BTCTestNetSyncer) Update() {
	assetSymbol := "BTC"
	for _, btcAccount := range btc.Account.EBalances[assetSymbol] {
		if btc.addressError[btcAccount.Address] {
			log.Println("Account info is not available at this moment, skip sync: ", btcAccount.Address)
			continue
		}
		herEthBalance := *big.NewInt(int64(0))
		storageKey := assetSymbol + "-" + btcAccount.Address
		if last, ok := btc.Storage.Get(btc.Account.Address); ok {
			// last-balance < External-ETH
			// Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
			if lastExtBalance, ok := last.LastExtBalance[storageKey]; ok && lastExtBalance != nil {
				// We need to guard here because buggy code before causing ext balance
				// for given btc account set to nil.
				if btc.ExtBalance[btcAccount.Address] == nil {
					btc.ExtBalance[btcAccount.Address] = big.NewInt(0)
				}
				if lastExtBalance.Cmp(btc.ExtBalance[btcAccount.Address]) < 0 {
					log.Printf("lastExtBalance.Cmp(btc.ExtBalance[%s])", btcAccount.Address)

					herEthBalance.Sub(btc.ExtBalance[btcAccount.Address], lastExtBalance)

					btcAccount.Balance += herEthBalance.Uint64()
					btcAccount.LastBlockHeight = btc.BlockHeight[btcAccount.Address].Uint64()
					btcAccount.Nonce = btc.Nonce[btcAccount.Address]
					btc.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, btc.ExtBalance[btcAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, btc.ExtBalance[btcAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(btc.Account)
					btc.Storage.Set(btc.Account.Address, last)

					log.Printf("New account balance after external balance credit: %v\n", last)
				}

				// last-balance < External-ETH
				// Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
				if lastExtBalance.Cmp(btc.ExtBalance[btcAccount.Address]) > 0 {
					log.Println("lastExtBalance.Cmp(btc.ExtBalance) ============")

					herEthBalance.Sub(lastExtBalance, btc.ExtBalance[btcAccount.Address])

					btcAccount.Balance -= herEthBalance.Uint64()
					btcAccount.LastBlockHeight = btc.BlockHeight[btcAccount.Address].Uint64()
					btcAccount.Nonce = btc.Nonce[btcAccount.Address]
					btc.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, btc.ExtBalance[btcAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, btc.ExtBalance[btcAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(btc.Account)
					btc.Storage.Set(btc.Account.Address, last)

					log.Printf("New account balance after external balance debit: %v\n", last)
				}
				continue
			}

			log.Printf("Initialise external balance in cache: %v\n", last)
			if btc.ExtBalance[btcAccount.Address] == nil {
				btc.ExtBalance[btcAccount.Address] = big.NewInt(0)
			}
			if btc.BlockHeight[btcAccount.Address] == nil {
				btc.BlockHeight[btcAccount.Address] = big.NewInt(0)
			}
			last = last.UpdateLastExtBalanceByKey(storageKey, btc.ExtBalance[btcAccount.Address])
			last = last.UpdateCurrentExtBalanceByKey(storageKey, btc.ExtBalance[btcAccount.Address])
			last = last.UpdateIsFirstEntryByKey(storageKey, true)
			last = last.UpdateIsNewAmountUpdateByKey(storageKey, false)
			btcAccount.UpdateBalance(btc.ExtBalance[btcAccount.Address].Uint64())
			btcAccount.UpdateBlockHeight(btc.BlockHeight[btcAccount.Address].Uint64())
			btcAccount.UpdateNonce(btc.Nonce[btcAccount.Address])
			btc.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount
			last = last.UpdateAccount(btc.Account)
			btc.Storage.Set(btc.Account.Address, last)
			continue
		}

		log.Printf("Initialise account in cache.")
		balance := btc.ExtBalance[btcAccount.Address]
		blockHeight := btc.BlockHeight[btcAccount.Address]
		lastbalances := make(map[string]*big.Int)
		lastbalances[storageKey] = btc.ExtBalance[btcAccount.Address]

		currentbalances := make(map[string]*big.Int)
		currentbalances[storageKey] = btc.ExtBalance[btcAccount.Address]
		if balance == nil {
			lastbalances[storageKey] = big.NewInt(0)
			currentbalances[storageKey] = big.NewInt(0)
		}
		isFirstEntry := make(map[string]bool)
		isFirstEntry[storageKey] = true
		isNewAmountUpdate := make(map[string]bool)
		isNewAmountUpdate[storageKey] = false
		if balance != nil {
			btcAccount.UpdateBalance(balance.Uint64())
		}
		if blockHeight != nil {
			btcAccount.UpdateBlockHeight(blockHeight.Uint64())
		}
		btcAccount.UpdateNonce(btc.Nonce[btcAccount.Address])

		btc.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount
		val := external.AccountCache{
			Account: btc.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
		}
		btc.Storage.Set(btc.Account.Address, val)
	}
}
