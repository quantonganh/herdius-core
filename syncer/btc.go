package sync

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"

	"github.com/herdius/herdius-core/p2p/log"
)

// BTCSyncer syncs all external BTC accounts.
type BTCSyncer struct {
	RPC    string
	syncer *ExternalSyncer
}

func newBTCSyncer() *BTCSyncer {
	b := &BTCSyncer{}
	b.syncer = newExternalSyncer("BTC")

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

	btcAccount, ok := btc.syncer.Account.EBalances[btc.syncer.assetSymbol]
	if !ok {
		return errors.New("BTC account does not exists")
	}

	apiKey := os.Getenv("BLOCKCHAIN_INFO_KEY")
	httpClient := newHTTPClient()

	for _, ba := range btcAccount {
		if len(apiKey) > 0 {
			url = btc.RPC + ba.Address + "?limit=1&api_code=" + apiKey
		} else {
			url = btc.RPC + ba.Address + "?limit=1"
		}
		resp, err := httpClient.Get(url)
		if err != nil {
			log.Error().Msgf("Error connecting Blockchain info: %v", err)
			btc.syncer.addressError[ba.Address] = true
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error().Msgf("failed to read response body: %v", err)
				btc.syncer.addressError[ba.Address] = true
				continue
			}

			var response BlockchainInfoResponse

			err = json.Unmarshal(bodyBytes, &response)
			if err != nil {
				log.Error().Msgf("failed to unmarshal response: %v", err)
				btc.syncer.addressError[ba.Address] = true
				continue
			}

			balance := big.NewInt(response.FinalBalance)
			// if tx>0 get last block height
			btc.syncer.BlockHeight[ba.Address] = big.NewInt(0)
			btc.syncer.Nonce[ba.Address] = uint64(0)

			if len(response.Txs) > 0 {
				if response.Txs[0].BlockHeight > 0 {
					btc.syncer.BlockHeight[ba.Address] = big.NewInt(int64(response.Txs[0].BlockHeight))
					btc.syncer.Nonce[ba.Address] = response.NTx

				}
			}

			btc.syncer.ExtBalance[ba.Address] = balance
			btc.syncer.addressError[ba.Address] = false
			continue
		}
		btc.syncer.addressError[ba.Address] = true
	}
	return nil

}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (btc *BTCSyncer) Update() {
	for _, assetAccount := range btc.syncer.Account.EBalances[btc.syncer.assetSymbol] {
		if btc.syncer.addressError[assetAccount.Address] {
			log.Warn().Msg("BTC Account info is not available at this moment, skip sync: " + assetAccount.Address)
			continue
		}
		btc.syncer.update(assetAccount.Address)
	}
}
