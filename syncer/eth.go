package sync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

// EthSyncer syncs all ETH external accounts
type EthSyncer struct {
	LastExtBalance map[string]*big.Int
	ExtBalance     map[string]*big.Int
	BlockHeight    map[string]*big.Int
	Nonce          map[string]uint64
	RPC            string
	Account        statedb.Account
	Storage        external.BalanceStorage
}

func newEthSyncer() *EthSyncer {
	e := &EthSyncer{}
	e.ExtBalance = make(map[string]*big.Int)
	e.LastExtBalance = make(map[string]*big.Int)
	e.BlockHeight = make(map[string]*big.Int)
	e.Nonce = make(map[string]uint64)

	return e
}

// GetExtBalance ...
func (es *EthSyncer) GetExtBalance() error {
	// If ETH account exists
	ethAccount, ok := es.Account.EBalances["ETH"]
	if !ok {
		return errors.New("ETH account does not exists")
	}

	for _, ea := range ethAccount {
		var (
			balance, latestBlockNumber *big.Int
			nonce                      uint64
			err                        error
		)
		client, err := ethclient.Dial(es.RPC)
		if err != nil {
			log.Println("Error connecting ETH RPC", err)
		}

		account := common.HexToAddress(ea.Address)

		// Get latest block number
		latestBlockNumber, err = es.getLatestBlockNumber(client)
		if err != nil {
			log.Println("Error getting ETH Latest block from RPC", err)
			return err
		}

		//Get nonce
		nonce, err = es.getNonce(client, account, latestBlockNumber)
		if err != nil {
			log.Println("Error getting ETH Account nonce from RPC", err)
			return err
		}

		balance, err = client.BalanceAt(context.Background(), account, latestBlockNumber)
		if err != nil {
			log.Println("Error getting ETH Balance from RPC", err)
			return err
		}
		es.ExtBalance[ea.Address] = balance
		es.BlockHeight[ea.Address] = latestBlockNumber
		es.Nonce[ea.Address] = nonce
	}

	return nil
}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (es *EthSyncer) Update() {
	assetSymbol := "ETH"
	for _, ethAccount := range es.Account.EBalances[assetSymbol] {
		fmt.Printf("%#v\n", ethAccount)
		herEthBalance := *big.NewInt(int64(0))
		storageKey := assetSymbol + "-" + ethAccount.Address
		if last, ok := es.Storage.Get(es.Account.Address); ok {
			// last-balance < External-ETH
			// Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
			if lastExtBalance, ok := last.LastExtBalance[storageKey]; ok {
				if lastExtBalance.Cmp(es.ExtBalance[ethAccount.Address]) < 0 {
					log.Printf("lastExtBalance.Cmp(es.ExtBalance[%s])", ethAccount.Address)
					herEthBalance.Sub(es.ExtBalance[ethAccount.Address], lastExtBalance)
					ethAccount.Balance += herEthBalance.Uint64()
					ethAccount.LastBlockHeight = es.BlockHeight[ethAccount.Address].Uint64()
					ethAccount.Nonce = es.Nonce[ethAccount.Address]
					es.Account.EBalances[assetSymbol][ethAccount.Address] = ethAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, es.ExtBalance[ethAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, es.ExtBalance[ethAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(es.Account)

					log.Printf("New account balance after external balance credit: %v\n", last)
					es.Storage.Set(es.Account.Address, last)
					continue

				}

				// last-balance < External-ETH
				// Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
				if lastExtBalance.Cmp(es.ExtBalance[ethAccount.Address]) > 0 {
					log.Println("lastExtBalance.Cmp(es.ExtBalance) ============")

					herEthBalance.Sub(lastExtBalance, es.ExtBalance[ethAccount.Address])
					ethAccount.Balance -= herEthBalance.Uint64()
					ethAccount.LastBlockHeight = es.BlockHeight[ethAccount.Address].Uint64()
					ethAccount.Nonce = es.Nonce[ethAccount.Address]
					last = last.UpdateLastExtBalanceByKey(storageKey, es.ExtBalance[ethAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, es.ExtBalance[ethAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					es.Account.EBalances[assetSymbol][ethAccount.Address] = ethAccount

					last = last.UpdateAccount(es.Account)

					log.Printf("New account balance after external balance debit: %v\n", last)
					es.Storage.Set(es.Account.Address, last)
				}
				continue
			}

			log.Printf("Initialise external balance in cache: %v\n", last)
			last = last.UpdateLastExtBalanceByKey(storageKey, es.ExtBalance[ethAccount.Address])
			last = last.UpdateCurrentExtBalanceByKey(storageKey, es.ExtBalance[ethAccount.Address])
			last = last.UpdateIsFirstEntryByKey(storageKey, true)
			last = last.UpdateIsNewAmountUpdateByKey(storageKey, false)
			ethAccount.UpdateBalance(es.ExtBalance[ethAccount.Address].Uint64())
			ethAccount.UpdateBlockHeight(es.BlockHeight[ethAccount.Address].Uint64())
			ethAccount.UpdateNonce(es.Nonce[ethAccount.Address])
			es.Account.EBalances[assetSymbol][ethAccount.Address] = ethAccount
			last = last.UpdateAccount(es.Account)
			es.Storage.Set(es.Account.Address, last)
			continue
		}
		balance := es.ExtBalance[ethAccount.Address]
		blockHeight := es.BlockHeight[ethAccount.Address]
		lastbalances := make(map[string]*big.Int)
		lastbalances[storageKey] = es.ExtBalance[ethAccount.Address]

		currentbalances := make(map[string]*big.Int)
		currentbalances[storageKey] = es.ExtBalance[ethAccount.Address]
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
		if blockHeight != nil {
			ethAccount.UpdateBlockHeight(blockHeight.Uint64())
		}
		ethAccount.UpdateNonce(es.Nonce[ethAccount.Address])

		es.Account.EBalances[assetSymbol][ethAccount.Address] = ethAccount
		val := external.AccountCache{
			Account: es.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
		}
		es.Storage.Set(es.Account.Address, val)
	}
}

func (es *EthSyncer) getLatestBlockNumber(client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
}

func (es *EthSyncer) getNonce(client *ethclient.Client, account common.Address, block *big.Int) (uint64, error) {
	nonce, err := client.NonceAt(context.Background(), account, block)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}
