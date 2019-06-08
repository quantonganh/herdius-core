package mempool

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/herdius/herdius-core/accounts/account"
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
	pending []mempoolTx
	queue   []mempoolTx
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
	tx     *protobuf.Tx
}

// Height returns the height for this transaction
func (memTx *mempoolTx) Height() int64 {
	return atomic.LoadInt64(&memTx.height)
}

// AddTx adds the tx Transaction to the MemPool and returns the total
// number of current Transactions within the MemPool
func (m *MemPool) AddTx(tx *protobuf.Tx) (int, int) {
	accSrv := account.NewAccountService()
	account, _ := accSrv.GetAccountByAddress(tx.GetSenderAddress())
	mpSize := len(m.pending)
	mt := mempoolTx{
		tx:     tx,
		height: int64(mpSize) + 1,
	}
	log.Println(tx.GetSenderAddress())

	log.Println(tx.GetAsset().Nonce)
	log.Println(account)

	if account != nil {
		if tx.GetAsset().Nonce == account.Nonce+1 {
			m.pending = append(m.pending, mt)
		}
	}
	m.queue = append(m.pending, mt)
	return len(m.pending), len(m.queue)
}

// GetTxs gets all transactions from the MemPool
func (m *MemPool) GetTxs() *tx.Txs {
	txs := &tx.Txs{}
	var cdc = amino.NewCodec()
	for _, mt := range m.pending {
		tx, _ := cdc.MarshalJSON(mt.tx)
		*txs = append(*txs, tx)
	}
	return txs
}

// GetTx returns a Tx for the given ID or nil if the corresponding TX exists
// Returns empty if Tx not found
func (m *MemPool) GetTx(id string) (int, *protobuf.Tx, error) {
	log.Println("Retrieving MemPool Tx's")
	for i, txQ := range m.pending {
		var cdc = amino.NewCodec()
		// txStr := &protobuf.Tx{}
		// err := cdc.UnmarshalJSON(txbz.tx, txStr)
		txbz, err := cdc.MarshalJSON(txQ.tx)
		if err != nil {
			return 0, nil, fmt.Errorf("unable to unmarshal tx bytes to txStr: %v", err)
		}

		txbzID := common.CreateTxID(txbz)
		if txbzID == id {
			log.Println("Matching transaction found for Tx ID:", id)
			return i, txQ.tx, nil
		}
	}
	return 0, nil, nil
}

// UpdateTx receives a Tx (newTx) and updates the corresponding Tx (origTx)
// with all non-empty fields in newTx
func (m *MemPool) UpdateTx(origI int, updated *protobuf.Tx) (*protobuf.Tx, error) {
	log.Println("Beginning update of transaction")
	//origBz := m.queue[origI].tx
	//var cdc = amino.NewCodec()
	orig := &protobuf.Tx{}
	// err := cdc.UnmarshalJSON(origBz, orig)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to unmarshal orig tx bytes to structured: %v", err)
	// }
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
	//updatedBz, err := cdc.MarshalJSON(orig)
	// if err != nil {
	// 	return nil, fmt.Errorf("could not marshal updated transaction back into memory pool: %v", err)
	// }

	m.pending[origI].tx = orig
	return orig, nil
}

// DeleteTx deletes a transaction currently in the MemPool by the transaction ID
// Returns true if successfully cancelled, false if can't find or cancel the transaction
func (m *MemPool) DeleteTx(id string) bool {
	log.Println("Beginning attempted removal from memory pool of Tx w/ ID:", id)
	for i, txStr := range m.pending {
		var cdc = amino.NewCodec()

		txbz, _ := cdc.MarshalJSON(txStr)
		mTxID := common.CreateTxID(txbz)
		if mTxID == id {
			log.Printf("Matched Tx ID (%v), removing from memory memory pool", id)
			m.pending = append(m.pending[:i], m.pending[i+1:]...)
			return true
		}
	}
	log.Printf("Unable to find Tx (id: %v) in memory pool", id)
	return false
}

// RemoveTxs transactions from the MemPool
func (m *MemPool) RemoveTxs(i int) {
	if len(m.pending) < i {
		m.pending = m.pending[len(m.pending):]
		return
	}
	m.pending = m.pending[i:]
}
