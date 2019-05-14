package mempool

import (
	"strconv"
	"testing"
	"fmt"
	"log"
	"encoding/json"
	"github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/supervisor/transaction"

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

func TestGetTxTrue(t *testing.T) {
	m := &MemPool{}
	nonce := "45"
	id := m.createTx(nonce)
	i, tx, err := m.GetTx(id)
	assert.Equal(t, 0, i)
	assert.Equal(t, fmt.Sprint(tx.Asset.Nonce), nonce)
	assert.NoError(t, err)
	nonce = "32"
	id = m.createTx(nonce)
	i, tx, err = m.GetTx(id)
	assert.Equal(t, 1, i)
	assert.Equal(t, fmt.Sprint(tx.Asset.Nonce), nonce)
	assert.NoError(t, err)
}
//
//func TestUpdateTxFalse(t *testing.T) {
//	// TODO
//}
//func TestUpdateTxTrue(t *testing.T) {
//	// TODO
//}

func (m *MemPool) createTx(i string) string {
	asset := &transaction.Asset{
		Category: "crypto",
		Symbol:   "HER",
		Network: "Herdius",
		Value:   "100",
		Fee:     "1",
		Nonce:   i,
	}

	tx := &transaction.Tx{
		SenderAddress:   "HDzLGL98C4vKtVWb3qzm92C2LX2V5kNhXR",
		SenderPubKey:    "A72fjBMhMkDgP+DQJOkPEngf76Xar99JqjgzGkEGjBWh",
		ReceiverAddress: "HPNMnZc9eNA7PzEMRWVqXwzPqieSRLzuyf",
		Asset:           *asset,
		Message:         "sending tokens",
	}
	
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	iint, _ := strconv.ParseInt(i, 64, 10)
	memPoolTx := mempoolTx{iint, txJSON}
	m.txs = append(m.txs, memPoolTx)
	
	txbzID := common.CreateTxID(txJSON)
	return txbzID
}