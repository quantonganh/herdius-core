package exbalance

import (
	"testing"

	"os"

	"github.com/stretchr/testify/assert"

	"github.com/herdius/herdius-core/storage/db"
)

type teststruct struct {
	Value string
}

func TestMemoryGetandSet(t *testing.T) {

	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	var m = NewDB(badgerdb)
	defer func() {
		m.Close()
		os.RemoveAll("./test.syncdb")
	}()
	key := "key"
	value := AccountCache{IsFirstHEREntry: true, IsNewHERAmountUpdate: true}
	m.Set(key, value)

	result, has := m.Get(key)
	assert.Equal(t, value.IsFirstHEREntry, result.IsFirstHEREntry, "Test byte comparision")
	assert.Equal(t, true, has, "Test has comparision")

}

func setup() db.DB {
	return LoadDB()
}
