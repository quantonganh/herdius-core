package exbalance

import (
	"fmt"

	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/storage/db"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

type memory struct {
	db db.DB
}

func init() {

	cryptoAmino.RegisterAmino(cdc)
}
func (b memory) Set(k string, x AccountCache) {
	by, _ := cdc.MarshalJSON(x)
	b.db.Set([]byte(k), by)
}

func (b memory) Get(k string) (AccountCache, bool) {
	var (
		v   AccountCache
		has bool
	)
	res := b.db.Get([]byte(k))
	e := cdc.UnmarshalJSON(res, &v)
	if e != nil {
		has = false
		return v, has
	}
	has = true
	return v, has

}

func (b *memory) GetAll() map[string]AccountCache {
	m := make(map[string]AccountCache)
	it, _ := b.db.BadgerIterator()
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		k := item.Key()
		v, _ := item.Value()
		var obj AccountCache
		cdc.UnmarshalJSON(v, &obj)
		m[string(k)] = obj
	}
	return m
}

func New() *memory {
	return &memory{db: LoadDB()}
}

func (m *memory) Close() {
	m.db.Close()
}
func NewDB(db db.DB) *memory {
	return &memory{db: db}
}

func LoadDB() db.DB {
	var dir string
	var dbName string
	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file not found...")
	} else {
		dir = viper.GetString("dev.syncdbpath")
		dbName = viper.GetString("dev.badgerDb")
	}

	return db.NewDB(dbName, db.GoBadgerBackend, dir)
}
