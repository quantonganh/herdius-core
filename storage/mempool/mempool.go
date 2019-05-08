package mempool

import (
	"sync"
	"sync/atomic"

	"github.com/herdius/herdius-core/tx"
)

// Service ...
type Service interface {
	AddTx(tx.Tx) int
	GetTxs() *tx.Txs
	RemoveTxs(int)
}

// MemPool ...
type MemPool struct {
	txs []mempoolTx
}

// Only one instance of MemPool will be instantiated.
var memPool *MemPool
var once sync.Once

// GetMemPool ..
func GetMemPool() *MemPool {
	once.Do(func() {
		memPool = &MemPool{}
	})
	return memPool
}

// mempoolTx is a transaction that successfully ran
type mempoolTx struct {
	height int64 // height that this tx had been validated in
	tx     tx.Tx
}

// Height returns the height for this transaction
func (memTx *mempoolTx) Height() int64 {
	return atomic.LoadInt64(&memTx.height)
}

// AddTx adds the tx Transaction to the MemPool and returns the total
// number of current Transactions within the MemPool
func (m *MemPool) AddTx(tx tx.Tx) int {
	mpSize := len(m.txs)

	mt := mempoolTx{
		tx:     tx,
		height: int64(mpSize) + 1,
	}
	m.txs = append(m.txs, mt)
	return len(m.txs)
}

// GetTxs gets all transactions from the MemPool
func (m *MemPool) GetTxs() *tx.Txs {
	txs := &tx.Txs{}
	for _, mt := range m.txs {
		*txs = append(*txs, mt.tx)
	}
	return txs
}

// GetTx returns a Tx for the given ID
func (m *MemPool) GetTx(id string) (*tx.Tx, error) {
	for _, txbz := range m.txs {
		// unmarshal tx-bytes into txstr
		txStr := cdc.Unmars(txs)

		// check txstr.id against id

	}

	return nil, nil
}

// UpdateTx receives a Tx (newTx) and updates the corresponding Tx (origTx)
// with all non-empty fields in newTx
func (m *MemPool) UpdateTx(tx *tx.Tx) *tx.Tx {
	return nil
}

// RemoveTxs transactions from the MemPool
func (m *MemPool) RemoveTxs(i int) {
	if len(m.txs) < i {
		m.txs = m.txs[len(m.txs):]
		return
	}
	m.txs = m.txs[i:]
}
