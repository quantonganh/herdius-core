package sync

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/storage/cache"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

func Sync(cache *cache.Cache) {
	blockchainSvc := &blockchain.Service{}
	//TODO: get latest block
	lastBlock, _ := blockchainSvc.GetBlockByHeight(int64(14))
	stateRoot := lastBlock.GetHeader().GetStateRoot()

	stateTrie, err := ethtrie.New(common.BytesToHash(stateRoot), statedb.GetDB())
	if err != nil {
		fmt.Errorf(fmt.Sprintf("Failed to retrieve the state trie: %v.", err))
		return
	}

	it := ethtrie.NewIterator(stateTrie.NodeIterator(stateRoot))

	for it.Next() {
		var senderAccount statedb.Account
		senderAddressBytes := it.Key
		senderActbz, err := stateTrie.TryGet(senderAddressBytes)
		if err != nil {
			fmt.Println("Failed to retrieve account detail: %v", err)
			continue
		}

		if len(senderActbz) > 0 {
			err = cdc.UnmarshalJSON(senderActbz, &senderAccount)
			if err != nil {
				fmt.Println("Failed to Unmarshal account: %v", err)
				continue
			}
		}
		es := &EthSyncer{Account: senderAccount, Cache: cache}
		es.GetExtBalance()
		es.Update()

	}

}
