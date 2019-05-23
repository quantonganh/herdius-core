package sync

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/storage/cache"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/spf13/viper"
)

func SyncAllAccounts(cache *cache.Cache) {
	var ethrpc string
	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file not found...")
	} else {
		ethrpc = viper.GetString("dev.ethrpc")
	}
	for {
		sync(cache, ethrpc)
	}
}

func sync(cache *cache.Cache, ethrpc string) {
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

		senderAccount.Erc20Address = "0xe2B5fcb42Bd8279a23Dc05fB4937E9BAf4d0D3Ab"

		es = &ERC20{Account: senderAccount, Cache: cache, RPC: ethrpc}
		es.GetExtBalance()
		es.Update()

	}

}
