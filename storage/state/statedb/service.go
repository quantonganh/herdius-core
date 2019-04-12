package statedb

import (
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	RegisterTrieAmino(cdc)
}

//RegisterTrieAmino ...
func RegisterTrieAmino(cdc *amino.Codec) {
	cryptoAmino.RegisterAmino(cdc)

}

var once sync.Once

// Trie modified referring Ethereum Merkle Patricia Trie.
type Trie interface {
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
	Commit(onleaf trie.LeafCallback) ([]byte, error)
	Hash() []byte //common.Hash
	NodeIterator(startKey []byte) trie.NodeIterator
	GetKey([]byte) []byte
}

// StateDB ...
// It has two files that are trie and the database where the KV pairs will be stored.
type state struct {
	trie *trie.Trie
	db   *trie.Database
}

// GetState :
// once.Do function ensures that the singleton is only instantiated once
func GetState(dir string) Trie {
	log.Println("dir:", dir)
	once.Do(func() {
		_, db := createGoLevelDB(dir)
		t, _ := trie.New(common.Hash{}, db)
		singleton = &state{trie: t, db: db}
	})
	return singleton
}

var singleton *state

func (s *state) GetTrie() *trie.Trie {
	return s.trie
}

//GetDB ...
func GetDB() *trie.Database {
	return singleton.db
}

func (s *state) TryGet(key []byte) ([]byte, error) {
	t := s.trie
	value, err := t.TryGet(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *state) TryUpdate(key, value []byte) error {
	t := s.trie

	err := t.TryUpdate(key, value)

	if err != nil {
		return err
	}
	return nil
}

func (s *state) TryDelete(key []byte) error {
	//t := s.trie

	// err := t.Delete(key)

	// if err != nil {
	// 	return err
	// }
	return nil
}

func (s *state) Commit(onleaf trie.LeafCallback) ([]byte, error) {
	t := s.trie
	root, err := t.Commit(nil)
	if err != nil {
		return nil, err
	}

	return root.Bytes(), nil
}

func (s *state) Hash() []byte {
	t := s.trie
	return t.Root()
}

func (s *state) NodeIterator(startKey []byte) trie.NodeIterator {
	return nil
}
func (s *state) GetKey([]byte) []byte {
	return nil
}

func createGoLevelDB(dir string) (string, *trie.Database) {
	diskdb, err := ethdb.NewLDBDatabase(dir, 0, 0)
	if err != nil {
		log.Fatalf(fmt.Sprintf("can't create state database: %v", err))
	}
	return dir, trie.NewDatabase(diskdb)
}
