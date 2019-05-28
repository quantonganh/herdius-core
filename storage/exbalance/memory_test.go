package exbalance

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/herdius/herdius-core/storage/db"
)

type teststruct struct {
	Value string
}

func TestMemoryGetandSet(t *testing.T) {

	var m = NewTest()
	defer m.CloseTest()
	key := "key"
	value := AccountCache{IsFirstHEREntry: true, IsNewHERAmountUpdate: true}
	//db := setup()
	m.Set(key, value)

	result, has := m.Get(key)
	assert.Equal(t, value.IsFirstHEREntry, result.IsFirstHEREntry, "Test byte comparision")
	assert.Equal(t, true, has, "Test has comparision")

	// cdc.UnmarshalJSON(result, &resl)
	//assert.Equal(t, value, result, "Test struct Comparison")

}

func setup() db.DB {
	return LoadDB()
}
