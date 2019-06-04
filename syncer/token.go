package sync

import (
	"context"
	"fmt"
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
	BlockHeight          *big.Int
	Nonce                uint64
	ExBal                external.BalanceStorage
	TokenContractAddress string
	TokenSymbol          string
	RPC                  string
}

//GetExtBalance Gets Asset balance from main chain
func (es *HERToken) GetExtBalance() error {
	var (
		latestBlockNumber *big.Int
		nonce             uint64
		err               error
	)
	client, err := ethclient.Dial(es.RPC)
	if err != nil {
		log.Println("Error connecting ETH RPC", err)
		return err
	}
	tokenAddress := common.HexToAddress(es.TokenContractAddress)
	address := common.HexToAddress(es.Account.Erc20Address)

	// Get latest block number
	latestBlockNumber, err = es.getLatestBlockNumber(client)
	if err != nil {
		log.Println("Error getting TOKEN Latest block from RPC", err)
		return err
	}

	//Get nonce
	nonce, err = es.getNonce(client, address, latestBlockNumber)
	if err != nil {
		log.Println("Error getting TOKEN Account nonce from RPC", err)
		return err
	}

	instance, err := contract.NewToken(tokenAddress, client)
	if err != nil {
		return err
	}
	bal, err := instance.BalanceOf(&bind.CallOpts{BlockNumber: latestBlockNumber}, address)
	if err != nil {
		return err
	}

	es.ExtBalance = bal
	es.BlockHeight = latestBlockNumber
	es.Nonce = nonce

	return nil
}

//Update Updates balance of asset in cache
func (es *HERToken) Update() {
	herBalance := *big.NewInt(int64(0))
	last, ok := es.ExBal.Get(es.Account.Address)

	if ok {
		fmt.Println("---last.LastExtHERBalance-------", last.LastExtHERBalance)
		//last-balance < External-ETH
		//Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
		lastExtHERBalance := last.LastExtHERBalance
		if lastExtHERBalance != nil {
			if lastExtHERBalance.Cmp(es.ExtBalance) < 0 {
				herBalance.Sub(es.ExtBalance, lastExtHERBalance)
				es.Account.Balance += herBalance.Uint64()
				es.Account.Nonce = es.Nonce
				es.Account.LastBlockHeight = es.BlockHeight.Uint64()

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
				es.Account.Nonce = es.Nonce
				es.Account.LastBlockHeight = es.BlockHeight.Uint64()
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
			es.Account.Nonce = es.Nonce
			es.Account.LastBlockHeight = es.BlockHeight.Uint64()

			last = last.UpdateAccount(es.Account)
			last = last.UpdateIsFirstHER(true)
			last = last.UpdateLastExtHERBalance(es.ExtBalance)
			last = last.UpdateCurrentExtHERBalance(es.ExtBalance)

			es.ExBal.Set(es.Account.Address, last)

		}

	} else {
		if len(es.Account.Erc20Address) > 0 {
			val := external.AccountCache{
				Account: es.Account,
			}
			es.ExBal.Set(es.Account.Address, val)
		}
	}

}

func (es *HERToken) getLatestBlockNumber(client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
}

func (es *HERToken) getNonce(client *ethclient.Client, account common.Address, block *big.Int) (uint64, error) {
	nonce, err := client.NonceAt(context.Background(), account, block)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}
