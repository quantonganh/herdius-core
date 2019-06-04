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

type BTCSyncer struct {
	LastExtBalance *big.Int
	ExtBalance     *big.Int
	Account        statedb.Account
	BlockHeight    *big.Int
	Nonce          uint64
	ExBal          external.BalanceStorage
	RPC            string
}

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

func (es *BTCSyncer) GetExtBalance() error {
	var url string

	btcAccount, ok := es.Account.EBalances["BTC"]
	if !ok {
		return errors.New("BTC account does not exists")
	}

	apiKey := os.Getenv("BLOCKCHAIN_INFO_KEY")
	if len(apiKey) > 0 {
		url = es.RPC + btcAccount.Address + "?limit=1&api_code=" + apiKey
	} else {
		url = es.RPC + btcAccount.Address + "?limit=1"
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error connecting Blockchain info ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var response BlockchainInfoResponse

		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return err
		}

		balance := big.NewInt(response.FinalBalance)
		//if tx>0 get last block height
		es.BlockHeight = big.NewInt(0)
		es.Nonce = uint64(0)

		if len(response.Txs) > 0 {
			es.BlockHeight = big.NewInt(int64(response.Txs[0].BlockHeight))
			es.Nonce = response.NTx
		}

		es.ExtBalance = balance
		return nil

	}
	return errors.New("Error getting external BTC balance")

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
					log.Println("Last balance is greater that external for asset", assetSymbol)

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
