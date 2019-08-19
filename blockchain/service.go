package blockchain

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/spf13/viper"

	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/crypto/herhash"
	pluginproto "github.com/herdius/herdius-core/hbi/protobuf"
	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/p2p/key"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/storage/db"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/herdius/herdius-core/supervisor/transaction"
)

// ServiceI is blockchain service interface
type ServiceI interface {
	GetBlockByHeight(height int64) (*protobuf.BaseBlock, error)
	CreateOrLoadGenesisBlock() (*protobuf.BaseBlock, error)
	GetBlockByBlockHash(db db.DB, key []byte) (*protobuf.BaseBlock, error)
	AddBaseBlock(bb *protobuf.BaseBlock) error
	GetLastBlock() *protobuf.BaseBlock
	GetTx(txID string) ([]byte, error)
}

// Service ...
type Service struct{}

var (
	_ ServiceI = (*Service)(nil)
)

// GetBlockByBlockHash ...
func (s *Service) GetBlockByBlockHash(db db.DB, key []byte) (*protobuf.BaseBlock, error) {
	bbbz := db.Get(key)

	bb := &protobuf.BaseBlock{}
	err := cdc.UnmarshalJSON(bbbz, bb)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal Base Block: %v", err)
	}

	return bb, nil
}

// CreateOrLoadGenesisBlock ...
func (s *Service) CreateOrLoadGenesisBlock() (*protobuf.BaseBlock, error) {
	genesisBlock, err := s.GetBlockByHeight(0)

	if err != nil {
		return nil, fmt.Errorf("failed while looking for the block: %v", err)
	}

	if genesisBlock != nil {
		blockHash := genesisBlock.GetHeader().GetBlock_ID().GetBlockHash()
		if blockHash != nil && len(blockHash) > 0 {
			return genesisBlock, nil
		}
	}

	// Create Genesis block

	ts := time.Now().UTC()

	timestamp := protobuf.Timestamp{
		Seconds: ts.Unix(),
		Nanos:   ts.UnixNano(),
	}

	blockID := &protobuf.BlockID{
		BlockHash: []byte{0},
	}

	// Get the initial state root for genesis block
	root, err := s.LoadStateDBWithInitialAccounts()

	if err != nil {
		return nil, fmt.Errorf("failed to create state root: %v", err)
	}
	header := &protobuf.BaseHeader{
		Time:      &timestamp,
		Height:    0,
		Block_ID:  blockID,
		StateRoot: root,
	}

	blockIDBz, err := cdc.MarshalJSON(blockID)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal Block ID: %v", err)
	}
	blockhash := herhash.Sum(blockIDBz)
	header.Block_ID.BlockHash = blockhash

	genesisBlock = &protobuf.BaseBlock{
		Header: header,
	}

	gbbz, err := cdc.MarshalJSON(genesisBlock)

	if err != nil {
		return nil, fmt.Errorf("failed to Marshal Base Block: %v", err)
	}
	badgerDB.Set(blockhash, gbbz)
	badgerDB.Set([]byte("LastBlock"), gbbz)

	return genesisBlock, nil
}

// getBlockByHeight query from blockHeightHashDB first, for O(1).
func (s *Service) getBlockByHeight(height int64) (*protobuf.BaseBlock, error) {
	var blockhash []byte
	err := blockHeightHashDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		var (
			err  error
			item *badger.Item
		)
		item, err = txn.Get([]byte(strconv.FormatInt(height, 10)))
		if err != nil {
			return err
		}

		blockhash, err = item.Value()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	baseBlock := &protobuf.BaseBlock{}
	err = badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get(blockhash)
		if err != nil {
			return err
		}

		v, err := item.Value()
		if err != nil {
			return err
		}
		if err := cdc.UnmarshalJSON(v, baseBlock); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return baseBlock, nil
}

// GetBlockByHeight ...
func (s *Service) GetBlockByHeight(height int64) (*protobuf.BaseBlock, error) {
	// Query from new DB first, for O(1) behavior. If any error, fall back to
	// O(n) behavior
	if baseBlock, err := s.getBlockByHeight(height); err == nil {
		return baseBlock, nil
	}

	lastBlock := &protobuf.BaseBlock{}
	lastBlockFlag := false
	err := badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}
			lb := &protobuf.BaseBlock{}
			err = cdc.UnmarshalJSON(v, lb)
			if err != nil {
				return nil
			}

			if height == lb.GetHeader().GetHeight() {
				lastBlock = lb
				lastBlockFlag = true
				return nil
			}

			if err != nil {
				return err
			}
			if lastBlockFlag {
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to find the block: %v.", err))
	}
	return lastBlock, nil
}

// AddBaseBlock adds base block to blockchain db
func (s *Service) AddBaseBlock(bb *protobuf.BaseBlock) error {
	blockhash := bb.GetHeader().GetBlock_ID().GetBlockHash()
	bbbz, err := cdc.MarshalJSON(bb)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to Marshal Block ID: %v.", err))
	}

	badgerDB.Set(blockhash, bbbz)
	badgerDB.Set([]byte("LastBlock"), bbbz)
	blockHeightHashDB.Set([]byte(strconv.FormatInt(bb.Header.GetHeight(), 10)), blockhash)
	return nil
}

// GetLastBlock ...
func (s *Service) GetLastBlock() *protobuf.BaseBlock {
	bbbz := badgerDB.Get([]byte("LastBlock"))
	bb := &protobuf.BaseBlock{}
	if len(bbbz) == 0 {
		bb, err := s.CreateOrLoadGenesisBlock()

		if err != nil {
			log.Error().Msgf("Error while creating genesis block: %v", err)

			return nil
		}
		return bb
	}

	cdc.UnmarshalJSON(bbbz, bb)
	return bb
}

// GetTx searches a transaction against a tx id in blockchain
// TODO: This will be completed in another PR since account
// since account registeration needs to be completed and
// I would like to complete the implementation once the
// account registeration works.
func (s *Service) GetTx(txID string) ([]byte, error) {
	err := badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}
			lb := &protobuf.BaseBlock{}
			err = cdc.UnmarshalJSON(v, lb)
			if err != nil {
				return nil
			}

			if lb.GetChildBlock() != nil && len(lb.GetChildBlock()) > 0 {
				// Check transactions in child blocks
			}

		}
		return nil
	})
	if err != nil {
		return []byte{0}, fmt.Errorf(fmt.Sprintf("Failed to find the tx: %v.", err))
	}
	return []byte{0}, nil
}

// TxServiceI is transaction service interface over blockchain
type TxServiceI interface {
	GetTx(id string) (*pluginproto.TxDetailResponse, error)
	GetTxs(address string) (*pluginproto.TxsResponse, error)
	GetTxsByAssetAndAddress(assetName, address string) (*pluginproto.TxsResponse, error)
	GetTxsByHeight(height int64) (*pluginproto.TxsResponse, error)
}

// TxService ...
type TxService struct{}

var (
	_ TxServiceI = (*TxService)(nil)
)

// GetTx ...
func (t *TxService) GetTx(id string) (*pluginproto.TxDetailResponse, error) {
	txDetailRes := &pluginproto.TxDetailResponse{}
	err := badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}

			var baseBlock protobuf.BaseBlock
			err = cdc.UnmarshalJSON(v, &baseBlock)
			if err != nil {
				return nil
			}

			// Check if base block has an transaction in it
			if baseBlock.GetTxsData() != nil &&
				len(baseBlock.GetTxsData().GetTx()) > 0 {

				// Get all the transaction from the base block
				txs := baseBlock.GetTxsData().GetTx()
				for _, txbz := range txs {
					var tx pluginproto.Tx
					err := cdc.UnmarshalJSON(txbz, &tx)
					if err != nil {
						log.Printf("Failed to Unmarshal tx: %v", err)
						continue
					}
					txID := getTxIDWithoutStatus(&tx)
					if txID == id {
						txDetailRes.Tx = &tx
						txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())

						ts := &pluginproto.Timestamp{
							Seconds: baseBlock.GetHeader().Time.Seconds,
							Nanos:   baseBlock.GetHeader().Time.Nanos,
						}
						txDetailRes.CreationDt = ts
						txDetailRes.TxId = txID
						break
					}
				}
			} else if len(baseBlock.GetChildBlock()) > 0 {
				var cbs []*protobuf.ChildBlock
				err := cdc.UnmarshalJSON(baseBlock.GetChildBlock(), &cbs)
				if err != nil {
					log.Printf("Failed to Unmarshal child block array: %v", err)
					continue
				}
				for _, cb := range cbs {
					// Get all the transaction from the child block
					txs := cb.GetTxsData().GetTx()
					for _, txbz := range txs {
						var txT transaction.Tx
						err := cdc.UnmarshalJSON(txbz, &txT)
						if err != nil {
							log.Printf("Failed to Unmarshal tx: %v", err)
							continue
						}
						tx, err := transactiontoProto(txT)
						if err != nil {
							log.Printf("Error converting Transation to Proto tx: %v", err)
							continue
						}
						txID := getTxIDWithoutStatus(&tx)
						if txID == id {
							txDetailRes.Tx = &tx
							txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())

							ts := &pluginproto.Timestamp{
								Seconds: baseBlock.GetHeader().Time.Seconds,
								Nanos:   baseBlock.GetHeader().Time.Nanos,
							}
							txDetailRes.CreationDt = ts
							txDetailRes.TxId = txID
							break
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Msgf("Failed to get blocks due to: %v", err.Error())
		return nil, fmt.Errorf("failed to get blocks due to: %v", err.Error())
	}
	return txDetailRes, nil
}

// GetTxs : Get all the txs by account address
func (t *TxService) GetTxs(address string) (*pluginproto.TxsResponse, error) {

	txDetails := make([]*pluginproto.TxDetailResponse, 0)

	err := badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		var duplicateTxTracker map[string]uint8
		duplicateTxTracker = make(map[string]uint8)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}

			var baseBlock protobuf.BaseBlock
			err = cdc.UnmarshalJSON(v, &baseBlock)
			if err != nil {
				return nil
			}

			// Check if base block has an transaction in it
			if baseBlock.GetTxsData() != nil &&
				len(baseBlock.GetTxsData().GetTx()) > 0 {

				// Get all the transaction from the base block
				txs := baseBlock.GetTxsData().GetTx()

				for _, txbz := range txs {
					var tx pluginproto.Tx
					err := cdc.UnmarshalJSON(txbz, &tx)
					if err != nil {
						log.Printf("Failed to Unmarshal tx: %v", err)
						continue
					}
					if strings.EqualFold(address, tx.SenderAddress) ||
						strings.EqualFold(address, tx.RecieverAddress) {
						txDetailRes := &pluginproto.TxDetailResponse{}
						txDetailRes.Tx = &tx
						txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())
						ts := &pluginproto.Timestamp{
							Seconds: baseBlock.GetHeader().Time.Seconds,
							Nanos:   baseBlock.GetHeader().Time.Nanos,
						}
						txDetailRes.CreationDt = ts
						txID := getTxIDWithoutStatus(&tx)
						txDetailRes.TxId = txID

						txDetails = append(txDetails, txDetailRes)
					}
				}
			} else if len(baseBlock.GetChildBlock()) > 0 {
				var cbs []*protobuf.ChildBlock
				err := cdc.UnmarshalJSON(baseBlock.GetChildBlock(), &cbs)
				if err != nil {
					log.Printf("Failed to Unmarshal child block array: %v", err)
					continue
				}
				for _, cb := range cbs {
					// Get all the transaction from the child block
					txs := cb.GetTxsData().GetTx()
					for _, txbz := range txs {
						var txT transaction.Tx
						err := cdc.UnmarshalJSON(txbz, &txT)
						if err != nil {
							log.Printf("Failed to Unmarshal tx: %v", err)
							continue
						}
						tx, err := transactiontoProto(txT)
						if err != nil {
							log.Printf("Error converting Transation to Proto tx: %v", err)
							continue
						}
						if strings.EqualFold(address, tx.SenderAddress) ||
							strings.EqualFold(address, tx.RecieverAddress) {
							txDetailRes := &pluginproto.TxDetailResponse{}
							txDetailRes.Tx = &tx
							txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())
							ts := &pluginproto.Timestamp{
								Seconds: baseBlock.GetHeader().Time.Seconds,
								Nanos:   baseBlock.GetHeader().Time.Nanos,
							}
							txDetailRes.CreationDt = ts
							txID := getTxIDWithoutStatus(&tx)
							if duplicateTxTracker[txID] == 1 {
								continue
							}
							duplicateTxTracker[txID] = 1
							txDetailRes.TxId = txID

							txDetails = append(txDetails, txDetailRes)
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Msgf("Failed to get blocks due to: %v", err.Error())
		return nil, fmt.Errorf("failed to get blocks due to: %v", err.Error())
	}
	txs := &pluginproto.TxsResponse{
		Txs: txDetails,
	}
	return txs, nil
}

// GetTxsByAssetAndAddress : Get all the txs by account address and asset name
func (t *TxService) GetTxsByAssetAndAddress(assetName, address string) (*pluginproto.TxsResponse, error) {
	txDetails := make([]*pluginproto.TxDetailResponse, 0)

	err := badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		var duplicateTxTracker map[string]uint8
		duplicateTxTracker = make(map[string]uint8)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}

			var baseBlock protobuf.BaseBlock
			err = cdc.UnmarshalJSON(v, &baseBlock)
			if err != nil {
				return nil
			}

			// Check if base block has an transaction in it
			if baseBlock.GetTxsData() != nil &&
				len(baseBlock.GetTxsData().GetTx()) > 0 {

				// Get all the transaction from the base block
				txs := baseBlock.GetTxsData().GetTx()

				for _, txbz := range txs {
					var tx pluginproto.Tx
					err := cdc.UnmarshalJSON(txbz, &tx)
					if err != nil {
						log.Printf("Failed to Unmarshal tx: %v", err)
						continue
					}
					if strings.EqualFold(assetName, tx.Asset.Symbol) {
						if strings.EqualFold(address, tx.SenderAddress) ||
							strings.EqualFold(address, tx.RecieverAddress) {
							txDetailRes := &pluginproto.TxDetailResponse{}
							txDetailRes.Tx = &tx
							txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())
							ts := &pluginproto.Timestamp{
								Seconds: baseBlock.GetHeader().Time.Seconds,
								Nanos:   baseBlock.GetHeader().Time.Nanos,
							}
							txDetailRes.CreationDt = ts
							txID := getTxIDWithoutStatus(&tx)
							//Remove duplicate transactions from the array
							//TODO: needs to be fixed
							if duplicateTxTracker[txID] == 1 {
								continue
							}
							duplicateTxTracker[txID] = 1
							txDetailRes.TxId = txID

							txDetails = append(txDetails, txDetailRes)
						}
					}

				}
			} else if len(baseBlock.GetChildBlock()) > 0 {
				var cbs []*protobuf.ChildBlock
				err := cdc.UnmarshalJSON(baseBlock.GetChildBlock(), &cbs)
				if err != nil {
					log.Printf("Failed to Unmarshal child block array: %v", err)
					continue
				}
				for _, cb := range cbs {
					// Get all the transaction from the child block
					txs := cb.GetTxsData().GetTx()
					for _, txbz := range txs {
						var txT transaction.Tx
						err := cdc.UnmarshalJSON(txbz, &txT)
						if err != nil {
							log.Printf("Failed to Unmarshal tx: %v", err)
							continue
						}
						tx, err := transactiontoProto(txT)
						if err != nil {
							log.Printf("Error converting Transation to Proto tx: %v", err)
							continue
						}
						if strings.EqualFold(assetName, tx.Asset.Symbol) {
							if strings.EqualFold(address, tx.SenderAddress) ||
								strings.EqualFold(address, tx.RecieverAddress) {
								txDetailRes := &pluginproto.TxDetailResponse{}
								txDetailRes.Tx = &tx
								txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())
								ts := &pluginproto.Timestamp{
									Seconds: baseBlock.GetHeader().Time.Seconds,
									Nanos:   baseBlock.GetHeader().Time.Nanos,
								}
								txDetailRes.CreationDt = ts
								txID := getTxIDWithoutStatus(&tx)
								//Remove duplicate transactions from the array
								//TODO: needs to be fixed
								if duplicateTxTracker[txID] == 1 {
									continue
								}
								duplicateTxTracker[txID] = 1
								txDetailRes.TxId = txID

								txDetails = append(txDetails, txDetailRes)
							}
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Msgf("Failed to get blocks due to: %v", err.Error())
		return nil, fmt.Errorf("failed to get blocks due to: %v", err.Error())
	}
	txs := &pluginproto.TxsResponse{
		Txs: txDetails,
	}
	return txs, nil
}

// GetLockedTxsByBlockNumber returns a list of all locked txs in a block
func (t *TxService) GetLockedTxsByBlockNumber(blockNumber int64) (*pluginproto.TxLockedResponse, error) {

	txs := make([]*pluginproto.TxDetailResponse, 0)

	err := badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}
			baseBlock := &protobuf.BaseBlock{}
			if err := cdc.UnmarshalJSON(v, baseBlock); err != nil {
				return err
			}

			if blockNumber == baseBlock.GetHeader().GetHeight() {
				// Check if base block has an transaction in it
				if baseBlock.GetTxsData() != nil {
					for _, txbz := range baseBlock.GetTxsData().GetTx() {
						var tx pluginproto.Tx
						err := cdc.UnmarshalJSON(txbz, &tx)
						if err != nil {
							return err
						}
						if strings.EqualFold("lock", tx.Type) {
							txDetailRes := &pluginproto.TxDetailResponse{}
							txDetailRes.Tx = &tx
							txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())
							ts := &pluginproto.Timestamp{
								Seconds: baseBlock.GetHeader().Time.Seconds,
								Nanos:   baseBlock.GetHeader().Time.Nanos,
							}
							txDetailRes.CreationDt = ts
							txDetailRes.TxId = getTxIDWithoutStatus(&tx)
							txs = append(txs, txDetailRes)
						}
					}
				}
				return nil
			}
		}
		return fmt.Errorf("block number %d not found", blockNumber)
	})
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to find the block: %v.", err))
	}

	return &pluginproto.TxLockedResponse{Txs: txs}, nil
}

// GetRedeemTxsByBlockNumber returns a list of all locked txs in a block
func (t *TxService) GetRedeemTxsByBlockNumber(blockNumber int64) (*pluginproto.TxRedeemResponse, error) {

	txs := make([]*pluginproto.TxDetailResponse, 0)

	err := badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}
			baseBlock := &protobuf.BaseBlock{}
			if err := cdc.UnmarshalJSON(v, baseBlock); err != nil {
				return err
			}

			if blockNumber == baseBlock.GetHeader().GetHeight() {
				// Check if base block has an transaction in it
				if baseBlock.GetTxsData() != nil {
					for _, txbz := range baseBlock.GetTxsData().GetTx() {
						var tx pluginproto.Tx
						err := cdc.UnmarshalJSON(txbz, &tx)
						if err != nil {
							return err
						}
						if strings.EqualFold("redeem", tx.Type) {
							txDetailRes := &pluginproto.TxDetailResponse{}
							txDetailRes.Tx = &tx
							txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())
							ts := &pluginproto.Timestamp{
								Seconds: baseBlock.GetHeader().Time.Seconds,
								Nanos:   baseBlock.GetHeader().Time.Nanos,
							}
							txDetailRes.CreationDt = ts
							txDetailRes.TxId = getTxIDWithoutStatus(&tx)
							txs = append(txs, txDetailRes)
						}
					}
				}
				return nil
			}
		}
		return fmt.Errorf("block number %d not found", blockNumber)
	})
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to find the block: %v.", err))
	}

	return &pluginproto.TxRedeemResponse{Txs: txs}, nil
}

// GetTxsByHeight returns a list of all txs in a given block by height
func (t *TxService) GetTxsByHeight(height int64) (*pluginproto.TxsResponse, error) {
	txs := make([]*pluginproto.TxDetailResponse, 0)
	var blockhash []byte

	err := blockHeightHashDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		var (
			err  error
			item *badger.Item
		)
		item, err = txn.Get([]byte(strconv.FormatInt(height, 10)))
		if err != nil {
			return err
		}

		blockhash, err = item.Value()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = badgerDB.GetBadgerDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get(blockhash)
		if err != nil {
			return err
		}

		v, err := item.Value()
		if err != nil {
			return err
		}
		baseBlock := &protobuf.BaseBlock{}
		if err := cdc.UnmarshalJSON(v, baseBlock); err != nil {
			return err
		}

		if baseBlock.GetTxsData() != nil {
			txs = make([]*pluginproto.TxDetailResponse, len(baseBlock.GetTxsData().GetTx()))
			for i, txbz := range baseBlock.GetTxsData().GetTx() {
				var tx pluginproto.Tx
				err := cdc.UnmarshalJSON(txbz, &tx)
				if err != nil {
					return err
				}
				txDetailRes := &pluginproto.TxDetailResponse{}
				txDetailRes.Tx = &tx
				txDetailRes.BlockId = uint64(baseBlock.GetHeader().GetHeight())
				ts := &pluginproto.Timestamp{
					Seconds: baseBlock.GetHeader().Time.Seconds,
					Nanos:   baseBlock.GetHeader().Time.Nanos,
				}
				txDetailRes.CreationDt = ts
				txDetailRes.TxId = getTxIDWithoutStatus(&tx)
				txs[i] = txDetailRes
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pluginproto.TxsResponse{Txs: txs}, nil
}

// LoadStateDBWithInitialAccounts loads state db with initial predefined accounts.
// Initially 50 accounts will be loaded to state db
func (s *Service) LoadStateDBWithInitialAccounts() ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var trie statedb.Trie
	var dir string
	viper.SetConfigName("config")       // Config file name without extension
	viper.AddConfigPath("../../config") // Path to config file
	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("config file not found: %v", err)
	}
	dir = viper.GetString("dev.statedbpath")

	trie = statedb.GetState(dir)
	parent := filepath.Dir(wd)
	for i := 0; i < 10; i++ {
		ser := i + 1
		filePath := filepath.Join(parent + "/herdius-core/cmd/testdata/secp205k1Accts/" + strconv.Itoa(ser) + "_peer_id.json")

		nodeKey, err := key.LoadOrGenNodeKey(filePath)

		if err != nil {
			return nil, fmt.Errorf("failed to Load or create node keys: %v", err)
		} else {
			pubKey := nodeKey.PrivKey.PubKey()
			b64PubKey := base64.StdEncoding.EncodeToString(pubKey.Bytes())
			account := statedb.Account{
				PublicKey: b64PubKey,
				Nonce:     0,
				Address:   pubKey.GetAddress(),
				Balance:   10000,
			}

			actbz, _ := cdc.MarshalJSON(account)
			pubKeyBytes := []byte(pubKey.GetAddress())
			err := trie.TryUpdate(pubKeyBytes, actbz)
			if err != nil {
				return nil, fmt.Errorf("failed to store account in state db: %v", err)
			}
		}
	}

	root, err := trie.Commit(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to commit the state trie: %v", err)
	}

	return root, nil
}

func GetBlockchainDb() db.DB {
	return badgerDB
}

// getTxIDWithoutStatus creates TxID without the status
func getTxIDWithoutStatus(tx *pluginproto.Tx) string {
	txWithOutStatus := *tx
	txWithOutStatus.Status = ""
	txbzWithOutStatus, _ := cdc.MarshalJSON(txWithOutStatus)
	txID := cmn.CreateTxID(txbzWithOutStatus)
	return txID
}

func transactiontoProto(txValue transaction.Tx) (tx pluginproto.Tx, err error) {
	val := uint64(0)
	if txValue.Asset.Value != "" {
		val, err = strconv.ParseUint(txValue.Asset.Value, 10, 64)
		if err != nil {
			err = fmt.Errorf("Failed to parse transaction value: %v", err)
		}
	}
	fee := uint64(0)
	if txValue.Asset.Fee != "" {
		fee, err = strconv.ParseUint(txValue.Asset.Fee, 10, 64)
		if err != nil {
			err = fmt.Errorf("Failed to parse transaction fee: %v", err)
		}
	}
	nonc := uint64(0)
	if txValue.Asset.Nonce != "" {
		nonc, err = strconv.ParseUint(txValue.Asset.Nonce, 10, 64)
		if err != nil {
			err = fmt.Errorf("Failed to parse transaction nonce: %v", err)
		}
	}
	asset := &pluginproto.Asset{
		Category:              txValue.Asset.Category,
		Symbol:                txValue.Asset.Symbol,
		Network:               txValue.Asset.Network,
		Value:                 val,
		Fee:                   fee,
		Nonce:                 nonc,
		ExternalSenderAddress: txValue.Asset.ExternalSenderAddress,
		LockedAmount:          txValue.Asset.LockedAmount,
		RedeemedAmount:        txValue.Asset.RedeemedAmount,
	}
	tx = pluginproto.Tx{
		SenderAddress:   txValue.SenderAddress,
		SenderPubkey:    txValue.SenderPubKey,
		RecieverAddress: txValue.ReceiverAddress,
		Asset:           asset,
		Message:         txValue.Message,
		Type:            txValue.Type,
		Sign:            txValue.Signature,
	}

	return
}
