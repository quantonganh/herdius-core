package statedb

import (
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
		var (
			err error
			ldb = new(ethdb.LDBDatabase)
			t   = new(trie.Trie)
		)
		ldb, err = loadLevelDB(dir)
		if err != nil {
			log.Fatalf("Error Loading LevelDB %v", err)
		}
		triedb := trie.NewDatabase(ldb)
		t, err = trie.New(common.Hash{}, triedb)
		if err != nil {
			log.Fatalf("Error Getting TrieDB %v", err)
		}
		singleton = &state{trie: t, db: triedb}
	})
	return singleton
}

//NewTrie
func NewTrie(hash common.Hash) (Trie, error) {
	state := new(state)
	t, err := trie.New(hash, singleton.db)
	if err != nil {
		return nil, err
	}
	state.db = singleton.db
	state.trie = t
	return state, nil
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
	s.db.Commit(root, true)

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

func loadLevelDB(dir string) (*ethdb.LDBDatabase, error) {
	return ethdb.NewLDBDatabase(dir, 0, 0)
}
