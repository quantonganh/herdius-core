package sync

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/storage/cache"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/spf13/viper"
)

func SyncAllAccounts(cache *cache.Cache) {
	var ethrpc, hercontractaddress string
	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file not found...")
	} else {
		infuraProjectID := os.Getenv("INFURAID")
		ethrpc = viper.GetString("dev.ethrpc")
		ethrpc = ethrpc + infuraProjectID
		log.Printf("Infura Url with Project ID: %v\n", ethrpc)
		hercontractaddress = viper.GetString("dev.hercontractaddress")

	}
	for {
		sync(cache, ethrpc, hercontractaddress)
	}
}

func sync(cache *cache.Cache, ethrpc, hercontractaddress string) {
	blockchainSvc := &blockchain.Service{}
	lastBlock := blockchainSvc.GetLastBlock()
	stateRoot := lastBlock.GetHeader().GetStateRoot()

	stateTrie, err := ethtrie.New(common.BytesToHash(stateRoot), statedb.GetDB())
	if err != nil {
		fmt.Printf("Failed to retrieve the state trie: %v.", err)
		return
	}
	it := ethtrie.NewIterator(stateTrie.NodeIterator(nil))

	for it.Next() {
		var senderAccount statedb.Account
		senderAddressBytes := it.Key
		senderActbz, err := stateTrie.TryGet(senderAddressBytes)
		if err != nil {
			fmt.Printf("Failed to retrieve account detail: %v", err)
			continue
		}

		if len(senderActbz) > 0 {
			err = cdc.UnmarshalJSON(senderActbz, &senderAccount)
			if err != nil {
				fmt.Printf("Failed to Unmarshal account: %v", err)
				continue
			}
		}
		var es Syncer
		es = &EthSyncer{Account: senderAccount, Cache: cache, RPC: ethrpc}
		es.GetExtBalance()
		es.Update()

		es = &HERToken{Account: senderAccount, Cache: cache, RPC: ethrpc, TokenContractAddress: hercontractaddress}
		es.GetExtBalance()
		es.Update()

	}

}
