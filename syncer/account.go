package sync

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/herdius/herdius-core/blockchain"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/spf13/viper"
)

type apiEndponts struct {
	btcRPC          string
	ethRPC          string
	herTokenAddress string
}

func SyncAllAccounts(exBal external.BalanceStorage) {
	var rpc apiEndponts
	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file not found...")
	} else {
		infuraProjectID := os.Getenv("INFURAID")
		rpc.ethRPC = viper.GetString("dev.ethrpc")
		rpc.ethRPC = rpc.ethRPC + infuraProjectID
		log.Printf("Infura Url with Project ID: %v\n", rpc.ethRPC)
		rpc.herTokenAddress = viper.GetString("dev.hercontractaddress")
		rpc.btcRPC = viper.GetString("dev.blockchaininforpc")

	}
	for {
		sync(exBal, rpc)
	}
}

func sync(exBal external.BalanceStorage, rpc apiEndponts) {
	blockchainSvc := &blockchain.Service{}
	lastBlock := blockchainSvc.GetLastBlock()
	stateRoot := lastBlock.GetHeader().GetStateRoot()

	stateTrie, err := ethtrie.New(common.BytesToHash(stateRoot), statedb.GetDB())
	if err != nil {
		log.Printf("Failed to retrieve the state trie: %v.", err)
		return
	}
	it := ethtrie.NewIterator(stateTrie.NodeIterator(nil))

	for it.Next() {
		var senderAccount statedb.Account
		senderAddressBytes := it.Key
		senderActbz, err := stateTrie.TryGet(senderAddressBytes)
		if err != nil {
			log.Printf("Failed to retrieve account detail: %v", err)
			continue
		}

		if len(senderActbz) > 0 {
			err = cdc.UnmarshalJSON(senderActbz, &senderAccount)
			if err != nil {
				log.Printf("Failed to Unmarshal account: %v", err)
				// Try unmarshal to old account struct
				var oldAccount statedb.OldAccount
				if err := cdc.UnmarshalJSON(senderActbz, &oldAccount); err != nil {
					log.Printf("Failed to Unmarshal old account: %v", err)
					continue
				}
				log.Printf("Sync old account before supporting multiple ebalances added.")
				senderAccount.Address = oldAccount.Address
				senderAccount.AddressHash = oldAccount.AddressHash
				senderAccount.Balance = oldAccount.Balance
				senderAccount.Erc20Address = oldAccount.Erc20Address
				senderAccount.ExternalNonce = oldAccount.ExternalNonce
				senderAccount.LastBlockHeight = oldAccount.LastBlockHeight
				senderAccount.Nonce = oldAccount.Nonce
				senderAccount.PublicKey = oldAccount.PublicKey
				senderAccount.StateRoot = oldAccount.StateRoot
				senderAccount.FirstExternalAddress = make(map[string]string)
				senderAccount.EBalances = make(map[string]map[string]statedb.EBalance)
				for asset, assetAccount := range oldAccount.EBalances {
					senderAccount.EBalances[asset] = make(map[string]statedb.EBalance)
					senderAccount.EBalances[asset][assetAccount.Address] = assetAccount
					senderAccount.FirstExternalAddress = make(map[string]string)
					senderAccount.FirstExternalAddress[asset] = assetAccount.Address
				}
			}
		}
		var syncers []Syncer

		// ETH syncer
		ethSyncer := newEthSyncer()
		ethSyncer.Account = senderAccount
		ethSyncer.Storage = exBal
		ethSyncer.RPC = rpc.ethRPC
		syncers = append(syncers, ethSyncer)

		// BTC syncer
		btcSyncer := newBTCSyncer()
		btcSyncer.Account = senderAccount
		btcSyncer.Storage = exBal
		btcSyncer.RPC = rpc.btcRPC
		syncers = append(syncers, btcSyncer)

		// HERDIUS syncer
		syncers = append(syncers, &HERToken{Account: senderAccount, Storage: exBal, RPC: rpc.ethRPC, TokenContractAddress: rpc.herTokenAddress})

		for _, asset := range syncers {
			// Dont update account if no new value recieved from respective api calls
			if asset.GetExtBalance() == nil {
				asset.Update()
			}

		}
	}

}
