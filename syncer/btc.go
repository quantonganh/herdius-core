package sync

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"

	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

// BTCSyncer syncs all external BTC accounts.
type BTCSyncer struct {
	LastExtBalance map[string]*big.Int
	ExtBalance     map[string]*big.Int
	BlockHeight    map[string]*big.Int
	Nonce          map[string]uint64
	RPC            string
	Account        statedb.Account
	Storage        external.BalanceStorage
}

func newBTCSyncer() *BTCSyncer {
	b := &BTCSyncer{}
	b.ExtBalance = make(map[string]*big.Int)
	b.LastExtBalance = make(map[string]*big.Int)
	b.BlockHeight = make(map[string]*big.Int)
	b.Nonce = make(map[string]uint64)

	return b
}

// BlockchainInfoResponse ...
type BlockchainInfoResponse struct {
	Hash160       string `json:"hash160"`
	Address       string `json:"address"`
	NTx           uint64 `json:"n_tx"`
	TotalReceived int    `json:"total_received"`
	TotalSent     int    `json:"total_sent"`
	FinalBalance  int64  `json:"final_balance"`
	Txs           []struct {
		Ver    int `json:"ver"`
		Inputs []struct {
			Sequence int64  `json:"sequence"`
			Witness  string `json:"witness"`
			PrevOut  struct {
				Spent             bool `json:"spent"`
				SpendingOutpoints []struct {
					TxIndex int `json:"tx_index"`
					N       int `json:"n"`
				} `json:"spending_outpoints"`
				TxIndex int    `json:"tx_index"`
				Type    int    `json:"type"`
				Addr    string `json:"addr"`
				Value   int    `json:"value"`
				N       int    `json:"n"`
				Script  string `json:"script"`
			} `json:"prev_out"`
			Script string `json:"script"`
		} `json:"inputs"`
		Weight      int    `json:"weight"`
		BlockHeight int    `json:"block_height"`
		RelayedBy   string `json:"relayed_by"`
		Out         []struct {
			Spent   bool   `json:"spent"`
			TxIndex int    `json:"tx_index"`
			Type    int    `json:"type"`
			Addr    string `json:"addr"`
			Value   int    `json:"value"`
			N       int    `json:"n"`
			Script  string `json:"script"`
		} `json:"out"`
		LockTime   int    `json:"lock_time"`
		Result     int    `json:"result"`
		Size       int    `json:"size"`
		BlockIndex int    `json:"block_index"`
		Time       int    `json:"time"`
		TxIndex    int    `json:"tx_index"`
		VinSz      int    `json:"vin_sz"`
		Hash       string `json:"hash"`
		VoutSz     int    `json:"vout_sz"`
	} `json:"txs"`
}

// GetExtBalance ...
func (es *BTCSyncer) GetExtBalance() error {
	var url string

	btcAccount, ok := es.Account.EBalances["BTC"]
	if !ok {
		return errors.New("BTC account does not exists")
	}

	apiKey := os.Getenv("BLOCKCHAIN_INFO_KEY")
	lastErr := errors.New("error getting external BTC balance")

	for _, ba := range btcAccount {
		if len(apiKey) > 0 {
			url = es.RPC + ba.Address + "?limit=1&api_code=" + apiKey
		} else {
			url = es.RPC + ba.Address + "?limit=1"
		}
		resp, err := http.Get(url)
		if err != nil {
			log.Println("Error connecting Blockchain info ", err)
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				lastErr = err
				continue
			}

			var response BlockchainInfoResponse

			err = json.Unmarshal(bodyBytes, &response)
			if err != nil {
				lastErr = err
				continue
			}

			balance := big.NewInt(response.FinalBalance)
			// if tx>0 get last block height
			es.BlockHeight[ba.Address] = big.NewInt(0)
			es.Nonce[ba.Address] = uint64(0)

			if len(response.Txs) > 0 {
				if response.Txs[0].BlockHeight > 0 {
					es.BlockHeight[ba.Address] = big.NewInt(int64(response.Txs[0].BlockHeight))
					es.Nonce[ba.Address] = response.NTx

				}
			}

			es.ExtBalance[ba.Address] = balance
			continue
		}
	}
	return lastErr

}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (es *BTCSyncer) Update() {
	assetSymbol := "BTC"
	for _, btcAccount := range es.Account.EBalances[assetSymbol] {
		herEthBalance := *big.NewInt(int64(0))
		storageKey := assetSymbol + "-" + btcAccount.Address
		if last, ok := es.Storage.Get(es.Account.Address); ok {
			// last-balance < External-ETH
			// Balance of ETH in H = Balance of ETH in H + ( Current_External_Bal - last_External_Bal_In_Cache)
			if lastExtBalance, ok := last.LastExtBalance[storageKey]; ok {
				if lastExtBalance.Cmp(es.ExtBalance[btcAccount.Address]) < 0 {
					log.Println("Last balance is less that external for asset address: ", assetSymbol, btcAccount.Address)
					herEthBalance.Sub(es.ExtBalance[btcAccount.Address], lastExtBalance)
					btcAccount.Balance += herEthBalance.Uint64()
					btcAccount.LastBlockHeight = es.BlockHeight[btcAccount.Address].Uint64()
					btcAccount.Nonce = es.Nonce[btcAccount.Address]
					es.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, es.ExtBalance[btcAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, es.ExtBalance[btcAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(es.Account)

					log.Printf("New account balance after external balance credit: %v\n", last)
					es.Storage.Set(es.Account.Address, last)
					continue
				}

				// last-balance < External-ETH
				// Balance of ETH in H1 	= Balance of ETH in H - ( last_External_Bal_In_Cache - Current_External_Bal )
				if lastExtBalance.Cmp(es.ExtBalance[btcAccount.Address]) > 0 {
					log.Println("Last balance is greater that external for asset address: ", assetSymbol, btcAccount.Address)
					herEthBalance.Sub(lastExtBalance, es.ExtBalance[btcAccount.Address])
					btcAccount.Balance -= herEthBalance.Uint64()
					btcAccount.LastBlockHeight = es.BlockHeight[btcAccount.Address].Uint64()
					btcAccount.Nonce = es.Nonce[btcAccount.Address]
					last = last.UpdateLastExtBalanceByKey(storageKey, es.ExtBalance[btcAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, es.ExtBalance[btcAccount.Address])
					last = last.UpdateIsFirstEntryByKey(storageKey, false)
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					es.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount
					last = last.UpdateAccount(es.Account)

					log.Printf("New account balance after external balance debit: %v\n", last)
					es.Storage.Set(es.Account.Address, last)
				}
				continue
			}

			log.Printf("Initialise external balance in cache: %v\n", last)
			last = last.UpdateLastExtBalanceByKey(storageKey, es.ExtBalance[btcAccount.Address])
			last = last.UpdateCurrentExtBalanceByKey(storageKey, es.ExtBalance[btcAccount.Address])
			last = last.UpdateIsFirstEntryByKey(storageKey, true)
			last = last.UpdateIsNewAmountUpdateByKey(storageKey, false)
			btcAccount.UpdateBalance(es.ExtBalance[btcAccount.Address].Uint64())
			btcAccount.UpdateBlockHeight(es.BlockHeight[btcAccount.Address].Uint64())
			btcAccount.UpdateNonce(es.Nonce[btcAccount.Address])
			es.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount
			last = last.UpdateAccount(es.Account)
			es.Storage.Set(es.Account.Address, last)
			continue
		}

		balance := es.ExtBalance[btcAccount.Address]
		blockHeight := es.BlockHeight[btcAccount.Address]
		lastbalances := make(map[string]*big.Int)
		lastbalances[storageKey] = es.ExtBalance[btcAccount.Address]
		currentbalances := make(map[string]*big.Int)
		currentbalances[storageKey] = es.ExtBalance[btcAccount.Address]
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
		btcAccount.UpdateNonce(es.Nonce[btcAccount.Address])
		es.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount
		val := external.AccountCache{
			Account: es.Account, LastExtBalance: lastbalances, CurrentExtBalance: currentbalances, IsFirstEntry: isFirstEntry, IsNewAmountUpdate: isNewAmountUpdate,
		}
		es.Storage.Set(es.Account.Address, val)
	}
}
