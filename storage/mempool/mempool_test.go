package mempool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMemPool(t *testing.T) {
	m := GetMemPool()
	assert.Implements(t, (*Service)(nil), m)
}

func TestGetTxFalse(t *testing.T) {
	m := MemPool{}
	i, tx, err := m.GetTx("aaaaaa11111")
	assert.Zero(t, i)
	assert.Nil(t, tx)
	assert.NoError(t, err)

}

//func TestGetTxTrue(t *testing.T) {
//	m := MemPool{}
//
//	//TODO
//
//}
//
//func TestUpdateTxFalse(t *testing.T) {
//	// TODO
//}
//func TestUpdateTxTrue(t *testing.T) {
//	// TODO
//}