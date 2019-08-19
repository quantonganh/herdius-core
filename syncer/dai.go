package sync

import (
	"context"
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/herdius/herdius-core/syncer/contract"
)

// DAISyncer syncs all DAI external accounts
type DAISyncer struct {
	LastExtBalance map[string]*big.Int
	ExtBalance     map[string]*big.Int
	BlockHeight    map[string]*big.Int
	Nonce          map[string]uint64
	RPC            string
	Account        statedb.Account
	Storage        external.BalanceStorage
	addressError   map[string]bool
	TokenAddress   string
}

func newDaiSyncer() *DAISyncer {
	d := &DAISyncer{}
	d.ExtBalance = make(map[string]*big.Int)
	d.LastExtBalance = make(map[string]*big.Int)
	d.Nonce = make(map[string]uint64)

	d.BlockHeight = make(map[string]*big.Int)
	d.addressError = make(map[string]bool)

	return d
}

// GetExtBalance ...
func (dai *DAISyncer) GetExtBalance() error {

	var (
		latestBlockNumber *big.Int
		err               error
	)

	client, err := ethclient.Dial(dai.RPC)
	if err != nil {
		log.Println("Error connecting DAI RPC", err)
		return err
	}

	// If DAI account exists
	daiAccount, ok := dai.Account.EBalances["DAI"]
	if !ok {
		return errors.New("DAI account does not exists")
	}

	for _, ta := range daiAccount {
		tokenAddress := common.HexToAddress(dai.TokenAddress)
		address := common.HexToAddress(ta.Address)

		latestBlockNumber, err = dai.getLatestBlockNumber(client)
		if err != nil {
			log.Println("Error getting DAI Latest block from RPC", err)
			dai.addressError[ta.Address] = true
			continue
		}

		//Get nonce
		nonce, err := dai.getNonce(client, address, latestBlockNumber)
		if err != nil {
			log.Println("Error getting DAI Balance from RPC", err)
			dai.addressError[ta.Address] = true
			continue
		}

		instance, err := contract.NewToken(tokenAddress, client)
		if err != nil {
			return err
		}
		bal, err := instance.BalanceOf(&bind.CallOpts{BlockNumber: latestBlockNumber}, address)
		if err != nil {
			dai.addressError[ta.Address] = true
			continue
		}
		dai.ExtBalance[ta.Address] = bal
		dai.BlockHeight[ta.Address] = latestBlockNumber
		dai.Nonce[ta.Address] = nonce
		dai.addressError[ta.Address] = false
	}

	return nil
}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (ts *DAISyncer) Update() {
	assetSymbol := "DAI"
	for _, daiAccount := range ts.Account.EBalances[assetSymbol] {
		if ts.addressError[daiAccount.Address] {
			log.Println("Account info is not available at this moment, skip sync: ", daiAccount.Address)
			continue
		}
		herTezosBalance := *big.NewInt(int64(0))
		storageKey := assetSymbol + "-" + daiAccount.Address
		if last, ok := ts.Storage.Get(ts.Account.Address); ok {
			// last-balance < External-XTZ
			// Balance of XTZ in H = Balance of XTZ in H + ( Current_External_Bal - last_External_Bal_In_Cache)
			if lastExtBalance, ok := last.LastExtBalance[storageKey]; ok && lastExtBalance != nil {
				if lastExtBalance.Cmp(ts.ExtBalance[daiAccount.Address]) < 0 {
					log.Printf("lastExtBalance.Cmp(ts.ExtBalance[%s])", daiAccount.Address)

					herTezosBalance.Sub(ts.ExtBalance[daiAccount.Address], lastExtBalance)

					daiAccount.Balance += herTezosBalance.Uint64()
					daiAccount.LastBlockHeight = ts.BlockHeight[daiAccount.Address].Uint64()
					ts.Account.EBalances[assetSymbol][daiAccount.Address] = daiAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, ts.ExtBalance[daiAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, ts.ExtBalance[daiAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(ts.Account)
					ts.Storage.Set(ts.Account.Address, last)
					log.Printf("New account balance after external balance credit: %v\n", last)
				}

				// last-balance < External-XTZ
				// Balance of XTZ in H1 	= Balance of XTZ in H - ( last_External_Bal_In_Cache - Current_External_Bal )
				if lastExtBalance.Cmp(ts.ExtBalance[daiAccount.Address]) > 0 {
					log.Println("lastExtBalance.Cmp(ts.ExtBalance) ============")

					herTezosBalance.Sub(lastExtBalance, ts.ExtBalance[daiAccount.Address])

					daiAccount.Balance -= herTezosBalance.Uint64()
					daiAccount.LastBlockHeight = ts.BlockHeight[daiAccount.Address].Uint64()
					ts.Account.EBalances[assetSymbol][daiAccount.Address] = daiAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, ts.ExtBalance[daiAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, ts.ExtBalance[daiAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(ts.Account)
					ts.Storage.Set(ts.Account.Address, last)
					log.Printf("New account balance after external balance debit: %v\n", last)
				}
				continue
			}

			log.Printf("Initialise external balance in cache: %v\n", last)
			if ts.ExtBalance[daiAccount.Address] == nil {
				ts.ExtBalance[daiAccount.Address] = big.NewInt(0)
			}
			if ts.BlockHeight[daiAccount.Address] == nil {
				ts.BlockHeight[daiAccount.Address] = big.NewInt(0)
			}
			last = last.UpdateLastExtBalanceByKey(storageKey, ts.ExtBalance[daiAccount.Address])
			last = last.UpdateCurrentExtBalanceByKey(storageKey, ts.ExtBalance[daiAccount.Address])
			last = last.UpdateIsFirstEntryByKey(storageKey, true)
			last = last.UpdateIsNewAmountUpdateByKey(storageKey, false)
			daiAccount.UpdateBalance(ts.ExtBalance[daiAccount.Address].Uint64())
			daiAccount.UpdateBlockHeight(ts.BlockHeight[daiAccount.Address].Uint64())
			ts.Account.EBalances[assetSymbol][daiAccount.Address] = daiAccount
			last = last.UpdateAccount(ts.Account)
			ts.Storage.Set(ts.Account.Address, last)
			continue
		}

		log.Printf("Initialise account in cache.")
		balance := ts.ExtBalance[daiAccount.Address]
		blockHeight := ts.BlockHeight[daiAccount.Address]
		lastbalances := make(map[string]*big.Int)
		lastbalances[storageKey] = ts.ExtBalance[daiAccount.Address]

		currentbalances := make(map[string]*big.Int)
		currentbalances[storageKey] = ts.ExtBalance[daiAccount.Address]
		if balance == nil {
			lastbalances[storageKey] = big.NewInt(0)
			currentbalances[storageKey] = big.NewInt(0)
		}
		isFirstEntry := make(map[string]bool)
		isFirstEntry[storageKey] = true
		isNewAmountUpdate := make(map[string]bool)
		isNewAmountUpdate[storageKey] = false
		if balance != nil {
			daiAccount.UpdateBalance(balance.Uint64())
		}
		if blockHeight != nil {
			daiAccount.UpdateBlockHeight(blockHeight.Uint64())
		}

		ts.Account.EBalances[assetSymbol][daiAccount.Address] = daiAccount
		val := external.AccountCache{
			Account: ts.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
		}
		ts.Storage.Set(ts.Account.Address, val)
	}
}

func (dai *DAISyncer) getLatestBlockNumber(client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
}

func (dai *DAISyncer) getNonce(client *ethclient.Client, account common.Address, block *big.Int) (uint64, error) {
	nonce, err := client.NonceAt(context.Background(), account, block)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}
