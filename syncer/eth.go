package sync

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

// EthSyncer syncs all ETH external accounts
type EthSyncer struct {
	RPC    string
	syncer *ExternalSyncer
}

func newEthSyncer() *EthSyncer {
	e := &EthSyncer{}
	e.syncer = newExternalSyncer("ETH")

	return e
}

// GetExtBalance ...
func (es *EthSyncer) GetExtBalance() error {
	// If ETH account exists
	ethAccount, ok := es.syncer.Account.EBalances[es.syncer.assetSymbol]
	if !ok {
		return errors.New("ETH account does not exists")
	}

	for _, ea := range ethAccount {
		var (
			balance, latestBlockNumber *big.Int
			nonce                      uint64
			err                        error
		)
		client, err := ethclient.Dial(es.RPC)
		if err != nil {
			log.Error().Msgf("Error connecting ETH RPC: %v", err)
			es.syncer.addressError[ea.Address] = true
			continue
		}

		account := common.HexToAddress(ea.Address)

		// Get latest block number
		latestBlockNumber, err = es.getLatestBlockNumber(client)
		if err != nil {
			log.Error().Msgf("Error getting ETH Latest block from RPC: %v", err)
			es.syncer.addressError[ea.Address] = true
			continue
		}

		// Get nonce
		nonce, err = es.getNonce(client, account, latestBlockNumber)
		if err != nil {
			log.Error().Msgf("Error getting ETH Account nonce from RPC: %v", err)
			es.syncer.addressError[ea.Address] = true
			continue
		}

		balance, err = client.BalanceAt(context.Background(), account, latestBlockNumber)
		if err != nil {
			log.Error().Msgf("Error getting ETH Balance from RPC: %v", err)
			es.syncer.addressError[ea.Address] = true
			continue
		}
		es.syncer.ExtBalance[ea.Address] = balance
		es.syncer.BlockHeight[ea.Address] = latestBlockNumber
		es.syncer.Nonce[ea.Address] = nonce
		es.syncer.addressError[ea.Address] = false
	}

	return nil
}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (es *EthSyncer) Update() {
	for _, assetAccount := range es.syncer.Account.EBalances[es.syncer.assetSymbol] {
		if es.syncer.addressError[assetAccount.Address] {
			log.Warn().Msg("ETH Account info is not available at this moment, skip sync: " + assetAccount.Address)
			continue
		}
		es.syncer.update(assetAccount.Address)
	}
}

func (es *EthSyncer) getLatestBlockNumber(client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
}

func (es *EthSyncer) getNonce(client *ethclient.Client, account common.Address, block *big.Int) (uint64, error) {
	nonce, err := client.NonceAt(context.Background(), account, block)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}
