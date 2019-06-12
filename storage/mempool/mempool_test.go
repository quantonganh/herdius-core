package mempool

import (
	"encoding/json"
	"log"
	"testing"

	acc "github.com/herdius/herdius-core/accounts/protobuf"
	"github.com/herdius/herdius-core/hbi/protobuf"
	"github.com/herdius/herdius-core/libs/common"

	"github.com/stretchr/testify/assert"
)

func TestAddTxHighNonce(t *testing.T) {
	m := GetMemPool()
	m.queue = m.queue[:0]
	m.pending = m.pending[:0]

	as := new(mockAccountService)
	tx, _ := NewTx(uint64(2), "nonce1")
	pending, queue := m.AddTx(&tx, as)
	assert.Equal(t, 1, pending, "pending tx")
	assert.Equal(t, 0, queue, "queue tx")
}

func TestAddTxGapNonce(t *testing.T) {
	m := GetMemPool()
	m.queue = m.queue[:0]
	m.pending = m.pending[:0]

	as := new(mockAccountService)
	tx, _ := NewTx(uint64(8), "nonce1")
	pending, queue := m.AddTx(&tx, as)
	assert.Equal(t, 0, pending, "pending tx")
	assert.Equal(t, 1, queue, "queue tx")
}

func TestProcessQueueNoGap(t *testing.T) {
	m := GetMemPool()
	m.queue = m.queue[:0]
	m.pending = m.pending[:0]

	as := new(mockAccountService)
	tx, _ := NewTx(uint64(3), "nonce1")
	m.AddTx(&tx, as)
	tx, _ = NewTx(uint64(2), "nonce1")
	m.AddTx(&tx, as)

	m.processQueue(as)
	assert.Equal(t, 2, len(m.pending), "pending tx")
	assert.Equal(t, 0, len(m.queue), "queue tx")
}

func TestProcessQueueGap(t *testing.T) {
	m := GetMemPool()
	m.queue = m.queue[:0]
	m.pending = m.pending[:0]

	as := new(mockAccountService)
	tx, _ := NewTx(uint64(9), "nonce1")
	m.AddTx(&tx, as)
	tx, _ = NewTx(uint64(3), "nonce1")
	m.AddTx(&tx, as)

	m.processQueue(as)
	assert.Equal(t, 0, len(m.pending), "pending tx")
	assert.Equal(t, 2, len(m.queue), "queue tx")
}
func NewTx(i uint64, address string) (protobuf.Tx, string) {
	asset := &protobuf.Asset{
		Category: "crypto",
		Symbol:   "HER",
		Network:  "Herdius",
		Value:    100,
		Fee:      1,
		Nonce:    i,
	}

	tx := protobuf.Tx{
		SenderAddress:   address,
		SenderPubkey:    "A72fjBMhMkDgP+DQJOkPEngf76Xar99JqjgzGkEGjBWh",
		RecieverAddress: "HPNMnZc9eNA7PzEMRWVqXwzPqieSRLzuyf",
		Asset:           asset,
		Message:         "sending tokens",
	}

	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}

	txbzID := common.CreateTxID(txJSON)
	return tx, txbzID
}

type mockAccountService struct {
}

func (m *mockAccountService) GetAccountByAddress(address string) (*acc.Account, error) {

	switch address {
	case "nonce1":
		return &acc.Account{Nonce: 1}, nil
	}
	return nil, nil

}
