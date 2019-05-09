package mempool

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/herdius/herdius-core/hbi/protobuf"
	"github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/tx"
	"github.com/tendermint/go-amino"
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
func (m *MemPool) GetTx(id string) (int, *protobuf.Tx, error) {
	log.Println("Retrieving MemPool Tx's")
	for i, txbz := range m.txs {
		var cdc = amino.NewCodec()
		txStr := &protobuf.Tx{}
		err := cdc.UnmarshalJSON(txbz.tx, txStr)
		if err != nil {
			return 0, nil, fmt.Errorf("unable to unmarshal tx bytes to txStr: %v", err)
		}

		txbzID := common.CreateTxID(txbz.tx)
		if txbzID == id {
			log.Println("Matching transaction found for Tx ID:", id)
			return i, txStr, nil
		}
	}
	return 0, nil, nil
}

// UpdateTx receives a Tx (newTx) and updates the corresponding Tx (origTx)
// with all non-empty fields in newTx
func (m *MemPool) UpdateTx(origI int, updated *protobuf.Tx) (*protobuf.Tx, error) {
	log.Println("Beginning update of transaction")
	origBz := m.txs[origI].tx
	var cdc = amino.NewCodec()
	orig := &protobuf.Tx{}
	err := cdc.UnmarshalJSON(origBz, orig)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal orig tx bytes to structured: %v", err)
	}
	if updated.RecieverAddress != "" && updated.RecieverAddress != orig.RecieverAddress {
		orig.RecieverAddress = updated.RecieverAddress
		log.Println("updated receiver address")
	}
	if updated.Message != "" && updated.Message != orig.Message {
		orig.Message = updated.Message
		log.Println("updated message")
	}
	if updated.Asset != nil && updated.Asset.Fee != 0 && updated.Asset.Fee != orig.Asset.Fee {
		orig.Asset.Fee = updated.Asset.Fee
		log.Println("updated tx fee")
	}
	if updated.Asset != nil && updated.Asset.Value != 0 && updated.Asset.Value != orig.Asset.Value {
		log.Println("updated tx value")
		orig.Asset.Value = updated.Asset.Value
	}
	if updated.Asset != nil && updated.Asset.Category != "" && updated.Asset.Category != orig.Asset.Category {
		log.Println("updated tx category")
		orig.Asset.Category = updated.Asset.Category
	}
	if updated.Asset != nil && updated.Asset.Symbol != "" && updated.Asset.Symbol != orig.Asset.Symbol {
		log.Println("updated tx symbol")
		orig.Asset.Symbol = updated.Asset.Symbol
	}
	if updated.Asset != nil && updated.Asset.Network != "" && updated.Asset.Network != orig.Asset.Network {
		log.Println("updated tx network")
		orig.Asset.Network = updated.Asset.Network
	}
	if updated.Type != "" && updated.Type != orig.Type {
		log.Println("updated type")
		orig.Type = updated.Type
	}
	updatedBz, err := cdc.MarshalJSON(orig)
	if err != nil {
		return nil, fmt.Errorf("could not marshal updated transaction back into memory pool: %v", err)
	}

	m.txs[origI].tx = updatedBz
	return orig, nil
}

// RemoveTxs transactions from the MemPool
func (m *MemPool) RemoveTxs(i int) {
	if len(m.txs) < i {
		m.txs = m.txs[len(m.txs):]
		return
	}
	m.txs = m.txs[i:]
}
