package blockchain

import (
	"fmt"
	"time"

	"github.com/herdius/herdius-core/supervisor/service"

	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/crypto/herhash"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/storage/db"
)

// ServiceI is blockchain service interface
type ServiceI interface {
	GetBlockByHeight(hieght int64) (*protobuf.BaseBlock, error)
	CreateOrLoadGenesisBlock() (*protobuf.BaseBlock, error)
	GetBlockByBlockHash(db db.DB, key []byte) (*protobuf.BaseBlock, error)
	AddBaseBlock(bb *protobuf.BaseBlock) error
	GetLastBlock() *protobuf.BaseBlock
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

	it, err := badgerDB.BadgerIterator()

	lastBlock := &protobuf.BaseBlock{}
	lastBlockFlag := false

	for it.Rewind(); it.Valid(); it.Next() {
		// item := it.Item()

		// err := item.Value(func(v []byte) error {
		// 	lb := &protobuf.BaseBlock{}
		// 	err := cdc.UnmarshalJSON(v, lb)
		// 	if err != nil {
		// 		return nil
		// 	}

		// 	if height == lb.GetHeader().GetHeight() {
		// 		lastBlock = lb
		// 		lastBlockFlag = true
		// 		return nil
		// 	}

		// 	return nil
		// })

		// if err != nil {
		// 	return nil, err
		// }
		if lastBlockFlag {
			break
		}
	}

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
