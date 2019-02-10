package db

import (
	"github.com/dgraph-io/badger"
	cmn "github.com/herdius/herdius-core/libs/common"
)

func init() {
	dbCreator := func(name string, dir string) (DB, error) {
		return NewBadgerDB(dir, dir)
	}
	registerDBCreator(GoBadgerBackend, dbCreator, false)
}

var _ DB = (*BadgerDB)(nil)

// BadgerDB ...
type BadgerDB struct {
	db *badger.DB
}

// NewBadgerDB ...
func NewBadgerDB(valueDir string, dir string) (*BadgerDB, error) {
	opts := badger.DefaultOptions

	return NewBadgerDBWithOpts(valueDir, dir, opts)
}

// NewBadgerDBWithOpts ...
func NewBadgerDBWithOpts(valueDir string, dir string, opts badger.Options) (*BadgerDB, error) {
	opts.Dir = dir
	opts.ValueDir = valueDir

	db, err := badger.Open(opts)

	if err != nil {
		return nil, err
	}
	database := &BadgerDB{
		db: db,
	}
	return database, nil

}

// GetBadgerDB ...
func (db *BadgerDB) GetBadgerDB() *badger.DB {
	return db.db
}
func (db *BadgerDB) Get(key []byte) []byte {
	var value []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			//cmn.PanicCrisis(err)
			return err
		}
		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			//cmn.PanicCrisis(err)
			return err
		}
		value = valCopy
		return nil
	})

	if err != nil {
		//cmn.PanicCrisis(err)
		return nil
	}

	return value
}

func (db *BadgerDB) Has(key []byte) bool {
	return false
}

func (db *BadgerDB) Set(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	err := db.db.Update(func(txn *badger.Txn) error {
		db.db.Lock()
		err := txn.Set(key, value)
		db.db.Unlock()
		return err
	})
	if err != nil {
		cmn.PanicCrisis(err)
	}
}

func (db *BadgerDB) SetSync(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)

	err := db.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})

	if err != nil {
		cmn.PanicCrisis(err)
	}
}

func (db *BadgerDB) Delete([]byte) {

}

func (db *BadgerDB) DeleteSync([]byte) {

}

// BadgerIterator ...
func (db *BadgerDB) BadgerIterator() (*badger.Iterator, error) {

	var iterator *badger.Iterator
	err := db.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		iterator = txn.NewIterator(opt)
		defer iterator.Close()
		return nil
	})
	if err != nil {
		//cmn.PanicCrisis(err)
		return nil, err
	}
	return iterator, nil
}

func (db *BadgerDB) Iterator(start, end []byte) Iterator {
	// db.BadgerIterator()
	// err := db.db.View(func(txn *badger.Txn) error {
	// 	opt := badger.DefaultIteratorOptions
	// 	it := txn.NewIterator(opt)
	// 	defer it.Close()

	// 	for it.Rewind(); it.Valid(); it.Next() {
	// 		item := it.Item()
	// 		k := item.Key()
	// 		err := item.Value(func(v []byte) error {
	// 			fmt.Printf("key=%s, value=%s\n", k, v)
	// 			return nil
	// 		})
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// })

	// if err != nil {
	// 	//cmn.PanicCrisis(err)
	// 	return nil
	// }
	return nil
}

func (db *BadgerDB) Close() {
	err := db.db.Close()
	if err != nil {
		cmn.PanicCrisis(err)
	}
}

func (db *BadgerDB) Print() {

}
