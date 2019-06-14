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
	addressError   map[string]bool
}

func newBTCSyncer() *BTCSyncer {
	b := &BTCSyncer{}
	b.ExtBalance = make(map[string]*big.Int)
	b.LastExtBalance = make(map[string]*big.Int)
	b.BlockHeight = make(map[string]*big.Int)
	b.Nonce = make(map[string]uint64)
	b.addressError = make(map[string]bool)

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
func (btc *BTCSyncer) GetExtBalance() error {
	var url string

	btcAccount, ok := btc.Account.EBalances["BTC"]
	if !ok {
		return errors.New("BTC account does not exists")
	}

	apiKey := os.Getenv("BLOCKCHAIN_INFO_KEY")

	for _, ba := range btcAccount {
		if len(apiKey) > 0 {
			url = btc.RPC + ba.Address + "?limit=1&api_code=" + apiKey
		} else {
			url = btc.RPC + ba.Address + "?limit=1"
		}
		resp, err := http.Get(url)
		if err != nil {
			log.Println("Error connecting Blockchain info ", err)
			btc.addressError[ba.Address] = true
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				btc.addressError[ba.Address] = true
				continue
			}

			var response BlockchainInfoResponse

			err = json.Unmarshal(bodyBytes, &response)
			if err != nil {
				btc.addressError[ba.Address] = true
				continue
			}

			balance := big.NewInt(response.FinalBalance)
			// if tx>0 get last block height
			btc.BlockHeight[ba.Address] = big.NewInt(0)
			btc.Nonce[ba.Address] = uint64(0)

			if len(response.Txs) > 0 {
				if response.Txs[0].BlockHeight > 0 {
					btc.BlockHeight[ba.Address] = big.NewInt(int64(response.Txs[0].BlockHeight))
					btc.Nonce[ba.Address] = response.NTx

				}
			}

			btc.ExtBalance[ba.Address] = balance
			btc.addressError[ba.Address] = false
			continue
		}
		btc.addressError[ba.Address] = true
	}
	return nil

}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (btc *BTCSyncer) Update() {
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

				if lastExtBalance.Cmp(btc.ExtBalance[btcAccount.Address]) == 0 && lastExtBalance.Uint64() != btcAccount.Balance {
					btcAccount.Balance = lastExtBalance.Uint64()
					btcAccount.LastBlockHeight = btc.BlockHeight[btcAccount.Address].Uint64()
					btcAccount.Nonce = btc.Nonce[btcAccount.Address]
					btc.Account.EBalances[assetSymbol][btcAccount.Address] = btcAccount

					last = last.UpdateLastExtBalanceByKey(storageKey, btc.ExtBalance[btcAccount.Address])
					last = last.UpdateCurrentExtBalanceByKey(storageKey, btc.ExtBalance[btcAccount.Address])
					last = last.UpdateIsNewAmountUpdateByKey(storageKey, true)
					last = last.UpdateAccount(btc.Account)
					btc.Storage.Set(btc.Account.Address, last)
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
