package sync

import (
	"os"
	"strings"
	stdSync "sync"

	"github.com/ethereum/go-ethereum/common"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/spf13/viper"

	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/p2p/log"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

type apiEndponts struct {
	btcRPC          string
	ethRPC          string
	herTokenAddress string
	hbtcRPC         string
	tezosRPC        string
}

// SyncAllAccounts syncs all assets of available accounts.
func SyncAllAccounts(exBal external.BalanceStorage, env string) {
	var rpc apiEndponts
	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Msg("failed to read config file")
		return
	}

	rpc.ethRPC = viper.GetString(env + ".ethrpc")
	rpc.herTokenAddress = viper.GetString(env + ".hercontractaddress")
	rpc.btcRPC = viper.GetString(env + ".blockchaininforpc")
	rpc.hbtcRPC = viper.GetString(env + ".hbtcrpc")
	rpc.tezosRPC = viper.GetString(env + ".tezosrpc")

	if strings.Index(rpc.ethRPC, ".infura.io") > -1 {
		rpc.ethRPC += os.Getenv("INFURAID")
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
		log.Error().Err(err).Msg("failed to retrieve the state trie.")
		return
	}
	it := ethtrie.NewIterator(stateTrie.NodeIterator(nil))

	log.Debug().Msg("Sync account start")
	var wg stdSync.WaitGroup
	for it.Next() {
		var senderAccount statedb.Account
		senderAddressBytes := it.Key
		senderActbz, err := stateTrie.TryGet(senderAddressBytes)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account detail")
			continue
		}

		if len(senderActbz) > 0 {
			err = cdc.UnmarshalJSON(senderActbz, &senderAccount)
			if err != nil {
				log.Warn().Err(err).Msg("failed to Unmarshal account")
				// Try unmarshal to old account struct
				var oldAccount statedb.OldAccount
				if err := cdc.UnmarshalJSON(senderActbz, &oldAccount); err != nil {
					log.Error().Err(err).Msg("failed to Unmarshal old account")
					continue
				}
				log.Debug().Msg("Sync old account before supporting multiple ebalances added.")
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
		ethSyncer.RPC = rpc.ethRPC
		ethSyncer.syncer.Account = senderAccount
		ethSyncer.syncer.Storage = exBal
		syncers = append(syncers, ethSyncer)

		// BTC syncer
		btcSyncer := newBTCSyncer()
		btcSyncer.RPC = rpc.btcRPC
		btcSyncer.syncer.Account = senderAccount
		btcSyncer.syncer.Storage = exBal
		syncers = append(syncers, btcSyncer)

		// HBTC syncer
		hbtcSyncer := newHBTCSyncer()
		hbtcSyncer.RPC = rpc.hbtcRPC
		hbtcSyncer.syncer.Account = senderAccount
		hbtcSyncer.syncer.Storage = exBal
		syncers = append(syncers, hbtcSyncer)

		// HBTC testnetsyncer
		hbtctestSyncer := newBTCTestNetSyncer()
		hbtctestSyncer.Account = senderAccount
		hbtctestSyncer.Storage = exBal
		syncers = append(syncers, hbtctestSyncer)

		// HERDIUS syncer
		syncers = append(syncers, &HERToken{Account: senderAccount, Storage: exBal, RPC: rpc.ethRPC, TokenContractAddress: rpc.herTokenAddress})

		// TEZOS syncer
		tezosSyncer := newTezosSyncer()
		tezosSyncer.RPC = rpc.tezosRPC
		tezosSyncer.syncer.Account = senderAccount
		tezosSyncer.syncer.Storage = exBal
		syncers = append(syncers, tezosSyncer)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, asset := range syncers {
				// Dont update account if no new value received from respective api calls
				if asset.GetExtBalance() == nil {
					asset.Update()
				}

			}
		}()
	}
	wg.Wait()
	log.Debug().Msg("Sync account end")
}
