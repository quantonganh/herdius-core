package blockchain

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/crypto/herhash"
	pluginproto "github.com/herdius/herdius-core/hbi/protobuf"
	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/storage/db"
	"github.com/herdius/herdius-core/supervisor/service"
)

// ServiceI is blockchain service interface
type ServiceI interface {
	GetBlockByHeight(hieght int64) (*protobuf.BaseBlock, error)
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
		return nil, fmt.Errorf(fmt.Sprintf("Failed to Unmarshal Base Block: %v.", err))
	}

	return bb, nil
}

// CreateOrLoadGenesisBlock ...
func (s *Service) CreateOrLoadGenesisBlock() (*protobuf.BaseBlock, error) {
	genesisBlock, err := s.GetBlockByHeight(0)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed while looking for the block: %v.", err))
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
	root, err := service.LoadStateDBWithInitialAccounts()

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to create state root: %v.", err))
	}
	header := &protobuf.BaseHeader{
		Time:      &timestamp,
		Height:    0,
		Block_ID:  blockID,
		StateRoot: root,
	}

	blockIDBz, err := cdc.MarshalJSON(blockID)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to Marshal Block ID: %v.", err))
	}
	blockhash := herhash.Sum(blockIDBz)
	header.Block_ID.BlockHash = blockhash

	genesisBlock = &protobuf.BaseBlock{
		Header: header,
	}

	gbbz, err := cdc.MarshalJSON(genesisBlock)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to Marshal Base Block: %v.", err))
	}
	badgerDB.Set(blockhash, gbbz)
	badgerDB.Set([]byte("LastBlock"), gbbz)

	return genesisBlock, nil
}

// GetBlockByHeight ...
func (s *Service) GetBlockByHeight(height int64) (*protobuf.BaseBlock, error) {
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
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Msgf("Failed to get blocks due to: %v", err.Error())
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get blocks due to: %v.", err.Error()))
	}
	return txDetailRes, nil
}

// getTxIDWithoutStatus creates TxID without the status
func getTxIDWithoutStatus(tx *pluginproto.Tx) string {
	txWithOutStatus := pluginproto.Tx{}
	txWithOutStatus.SenderAddress = tx.SenderAddress
	txWithOutStatus.RecieverAddress = tx.RecieverAddress
	txWithOutStatus.Message = tx.Message
	txWithOutStatus.Sign = tx.Sign
	txWithOutStatus.Type = tx.Type
	txWithOutStatus.SenderPubkey = tx.SenderPubkey
	txWithOutStatus.Asset = tx.Asset
	txbzWithOutStatus, _ := cdc.MarshalJSON(txWithOutStatus)
	txID := cmn.CreateTxID(txbzWithOutStatus)
	return txID
}
