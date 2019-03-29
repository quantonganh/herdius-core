package mempool

import (
	"log"

	"github.com/herdius/herdius-core/tx"
)

type Service interface {
	AddTxs(tx.Service) (int, error)
}

type MemPool struct {
	TxRam tx.Txs
}

func NewMemPool() Service {
	return &MemPool{}
}

// AddTx adds the txb Transaction to the MemPool and returns to the total
// number of current Transactions within the MemPool
func (m *MemPool) AddTxs(txs tx.Service) (int, error) {
	for _, txb := range txs.GetTxs() {
		m.TxRam = append(m.TxRam, txb)
	}
	log.Println("mempool txcount:", len(m.TxRam))
	return len(m.TxRam), nil
}
