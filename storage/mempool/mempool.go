package mempool

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/herdius/herdius-core/tx"
)

// Service ...
type Service interface {
	//AddTxs(tx.Service) (int, error)
	AddTx(tx.Tx) int
	GetTxs() (tx.Txs, error)
	RemoveTxs()
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
func (m *MemPool) GetTxs() tx.Txs {
	txs := make(tx.Txs, 0)
	for _, mt := range m.txs {
		txs = append(txs, mt.tx)
	}
	return txs
}

// RemoveTxs transactions from the MemPool
func (m *MemPool) RemoveTxs() {
	// TODO: 500 needs to be configurable
	if len(m.txs) < 500 {
		m.txs = m.txs[len(m.txs):]
		log.Println("mempool txcount after remove:", len(m.txs))
		return
	}
	log.Println("mempool txcount after remove:", len(m.txs))
	m.txs = m.txs[500:]
}
