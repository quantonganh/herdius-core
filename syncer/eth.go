package sync

import (
	"context"
	"errors"
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
	BlockHeight    *big.Int
	Nonce          uint64
	Account        statedb.Account
	ExBal          external.BalanceStorage
	RPC            string
}

func (es *EthSyncer) GetExtBalance() error {
	var (
		balance, latestBlockNumber *big.Int
		nonce                      uint64
		err                        error
	)
	client, err := ethclient.Dial(es.RPC)
	if err != nil {
		log.Println("Error connecting ETH RPC", err)
	}
	// If ETH account exists
	ethAccount, ok := es.Account.EBalances["ETH"]
	if !ok {
		return errors.New("ETH account does not exists")
	}

	account := common.HexToAddress(ethAccount.Address)

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
	es.ExtBalance = balance
	es.BlockHeight = latestBlockNumber
	es.Nonce = nonce

	return nil
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
					value.LastBlockHeight = es.BlockHeight.Uint64()
					value.Nonce = es.Nonce
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
					value.LastBlockHeight = es.BlockHeight.Uint64()
					value.Nonce = es.Nonce
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
				value.UpdateBlockHeight(es.BlockHeight.Uint64())
				value.UpdateNonce(es.Nonce)
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
			value.UpdateBlockHeight(es.BlockHeight.Uint64())
			value.UpdateNonce(es.Nonce)

			es.Account.EBalances[assetSymbol] = value
			val := external.AccountCache{
				Account: es.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
			}
			es.ExBal.Set(es.Account.Address, val)
		}

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
