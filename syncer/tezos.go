package sync

import (
	"errors"
	"log"
	"math/big"

	goTezos "github.com/DefinitelyNotAGoat/go-tezos"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

// TezosSyncer syncs all XTZ external accounts
type TezosSyncer struct {
	LastExtBalance map[string]*big.Int
	ExtBalance     map[string]*big.Int
	BlockHeight    map[string]*big.Int
	RPC            string
	Account        statedb.Account
	Storage        external.BalanceStorage
	addressError   map[string]bool
}

func newTezosSyncer() *TezosSyncer {
	t := &TezosSyncer{}
	t.ExtBalance = make(map[string]*big.Int)
	t.LastExtBalance = make(map[string]*big.Int)
	t.BlockHeight = make(map[string]*big.Int)
	t.addressError = make(map[string]bool)

	return t
}

// GetExtBalance ...
func (ts *TezosSyncer) GetExtBalance() error {
	// If XTZ account exists
	xtsAccount, ok := ts.Account.EBalances["XTZ"]
	if !ok {
		return errors.New("XTZ account does not exists")
	}

	for _, ta := range xtsAccount {
		// TODO: remove empty argument when go-tezos fixes it.
		gt, err := goTezos.NewGoTezos(ts.RPC, "")
		if err != nil {
			log.Println("Error connecting XTZ RPC", err)
			ts.addressError[ta.Address] = true
			continue
		}

		// Get latest block number
		latestBlockNumber, err := ts.getLatestBlockNumber(gt)
		if err != nil {
			log.Println("Error getting XTZ Latest block from RPC", err)
			ts.addressError[ta.Address] = true
			continue
		}

		balance, err := gt.Account.GetBalanceAtBlock(ta.Address, latestBlockNumber)
		if err != nil {
			log.Println("Error getting XTZ Balance from RPC", err)
			ts.addressError[ta.Address] = true
			continue
		}
		ts.ExtBalance[ta.Address] = big.NewInt(int64(balance) * goTezos.MUTEZ)
		ts.BlockHeight[ta.Address] = latestBlockNumber
		ts.addressError[ta.Address] = false
	}

	return nil
}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (ts *TezosSyncer) Update() {
	assetSymbol := "XTZ"
	for _, xtsAccount := range ts.Account.EBalances[assetSymbol] {
		if ts.addressError[xtsAccount.Address] {
			log.Println("Account info is not available at this moment, skip sync: ", xtsAccount.Address)
			continue
		}
		herTezosBalance := *big.NewInt(int64(0))
		storageKey := assetSymbol + "-" + xtsAccount.Address
		if last, ok := ts.Storage.Get(ts.Account.Address); ok {
			// last-balance < External-XTZ
			// Balance of XTZ in H = Balance of XTZ in H + ( Current_External_Bal - last_External_Bal_In_Cache)
			if lastExtBalance, ok := last.LastExtBalance[storageKey]; ok && lastExtBalance != nil {
				if lastExtBalance.Cmp(ts.ExtBalance[xtsAccount.Address]) < 0 {
					log.Printf("lastExtBalance.Cmp(ts.ExtBalance[%s])", xtsAccount.Address)

					herTezosBalance.Sub(ts.ExtBalance[xtsAccount.Address], lastExtBalance)

					xtsAccount.Balance += herTezosBalance.Uint64()
					xtsAccount.LastBlockHeight = ts.BlockHeight[xtsAccount.Address].Uint64()
					ts.Account.EBalances[assetSymbol][xtsAccount.Address] = xtsAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, ts.ExtBalance[xtsAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, ts.ExtBalance[xtsAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(ts.Account)
					ts.Storage.Set(ts.Account.Address, last)
					log.Printf("New account balance after external balance credit: %v\n", last)
				}

				// last-balance < External-XTZ
				// Balance of XTZ in H1 	= Balance of XTZ in H - ( last_External_Bal_In_Cache - Current_External_Bal )
				if lastExtBalance.Cmp(ts.ExtBalance[xtsAccount.Address]) > 0 {
					log.Println("lastExtBalance.Cmp(ts.ExtBalance) ============")

					herTezosBalance.Sub(lastExtBalance, ts.ExtBalance[xtsAccount.Address])

					xtsAccount.Balance -= herTezosBalance.Uint64()
					xtsAccount.LastBlockHeight = ts.BlockHeight[xtsAccount.Address].Uint64()
					ts.Account.EBalances[assetSymbol][xtsAccount.Address] = xtsAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, ts.ExtBalance[xtsAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, ts.ExtBalance[xtsAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(ts.Account)
					ts.Storage.Set(ts.Account.Address, last)
					log.Printf("New account balance after external balance debit: %v\n", last)
				}
				continue
			}

			log.Printf("Initialise external balance in cache: %v\n", last)
			if ts.ExtBalance[xtsAccount.Address] == nil {
				ts.ExtBalance[xtsAccount.Address] = big.NewInt(0)
			}
			if ts.BlockHeight[xtsAccount.Address] == nil {
				ts.BlockHeight[xtsAccount.Address] = big.NewInt(0)
			}
			last = last.UpdateLastExtBalanceByKey(storageKey, ts.ExtBalance[xtsAccount.Address])
			last = last.UpdateCurrentExtBalanceByKey(storageKey, ts.ExtBalance[xtsAccount.Address])
			last = last.UpdateIsFirstEntryByKey(storageKey, true)
			last = last.UpdateIsNewAmountUpdateByKey(storageKey, false)
			xtsAccount.UpdateBalance(ts.ExtBalance[xtsAccount.Address].Uint64())
			xtsAccount.UpdateBlockHeight(ts.BlockHeight[xtsAccount.Address].Uint64())
			ts.Account.EBalances[assetSymbol][xtsAccount.Address] = xtsAccount
			last = last.UpdateAccount(ts.Account)
			ts.Storage.Set(ts.Account.Address, last)
			continue
		}

		log.Printf("Initialise account in cache.")
		balance := ts.ExtBalance[xtsAccount.Address]
		blockHeight := ts.BlockHeight[xtsAccount.Address]
		lastbalances := make(map[string]*big.Int)
		lastbalances[storageKey] = ts.ExtBalance[xtsAccount.Address]

		currentbalances := make(map[string]*big.Int)
		currentbalances[storageKey] = ts.ExtBalance[xtsAccount.Address]
		if balance == nil {
			lastbalances[storageKey] = big.NewInt(0)
			currentbalances[storageKey] = big.NewInt(0)
		}
		isFirstEntry := make(map[string]bool)
		isFirstEntry[storageKey] = true
		isNewAmountUpdate := make(map[string]bool)
		isNewAmountUpdate[storageKey] = false
		if balance != nil {
			xtsAccount.UpdateBalance(balance.Uint64())
		}
		if blockHeight != nil {
			xtsAccount.UpdateBlockHeight(blockHeight.Uint64())
		}

		ts.Account.EBalances[assetSymbol][xtsAccount.Address] = xtsAccount
		val := external.AccountCache{
			Account: ts.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
		}
		ts.Storage.Set(ts.Account.Address, val)
	}
}

func (ts *TezosSyncer) getLatestBlockNumber(client *goTezos.GoTezos) (*big.Int, error) {
	block, err := client.Block.GetHead()
	if err != nil {
		return nil, err
	}
	return big.NewInt(int64(block.Header.Level)), nil
}
