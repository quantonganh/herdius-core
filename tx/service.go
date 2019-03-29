package tx

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/crypto/herhash"
	"github.com/herdius/herdius-core/crypto/merkle"
	txProtoc "github.com/herdius/herdius-core/hbi/protobuf"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	RegisterTxServiceAmino(cdc)
}

//RegisterTxServiceAmino ...
func RegisterTxServiceAmino(cdc *amino.Codec) {
	cryptoAmino.RegisterAmino(cdc)

}

// Service ... Operations for a list of transactions
type Service interface {
	SetTxs(txs [][]byte)
	GetTxs() [][]byte
	ParseNewTxRequest(uint64, uint64, *txProtoc.TransactionRequest) error // Accepts incoming protoc structured transactions and sets byte-wise tx transaction
	Hash(tx Tx) []byte                                                    // Hash computes the hash of the wire encoded transaction.
	String(tx Tx) string                                                  // String returns the hex-encoded transaction as a string.
	Index(tx Tx) int                                                      // Index returns the index of this transaction in the list, or -1 if not found
	IndexByHash(hash []byte) int                                          // IndexByHash returns the index of this transaction hash in the list, or -1 if not found
	MerkleHash() []byte                                                   // MerkleHash returns the simple Merkle root hash of the transactions.
	Proof(i int) Proof                                                    // Proof returns a simple merkle proof for this node.
	LeafHash(tx Tx) []byte                                                // LeafHash returns the hash of the this proof refers to.
	// Validate verifies the proof. It returns nil if the RootHash matches the dataHash argument,
	// and if the proof is internally consistent. Otherwise, it returns a sensible error.
	Validate(dataHash []byte) error
}

// GetTxService creates a new transaction service
func GetTxsService() Service {
	return &Txs{}
}

func (t *Txs) ParseNewTxRequest(senderNonce, senderBal uint64, txReq *txProtoc.TransactionRequest) error {

	reqNonce := txReq.Tx.Asset.Nonce

	if senderNonce >= reqNonce {
		log.Println("request nonce must be larger than sender account's nonce value")
		log.Println("request nonce:", reqNonce)
		log.Println("sender account's nonce:", senderNonce)
		return errors.New("request nonce must be larger than sender account's nonce value")
	}

	tx, err := cdc.MarshalJSON(txReq)
	*t = append(*t, tx)
	if err != nil {
		log.Println("marshalling error:", err)
		return err
	}
	return nil
}

func (t *Txs) SetTxs(txs [][]byte) {
	*t = txs
}

func (t *Txs) GetTxs() [][]byte {
	var allTxs [][]byte
	for _, v := range *t {
		allTxs = append(allTxs, v)
	}
	return allTxs
}

// Hash ...
func (t *Txs) Hash(tx Tx) []byte {
	return herhash.Sum(tx)
}

// String returns the hex-encoded transaction as a string.
func (t *Txs) String(tx Tx) string {
	return fmt.Sprintf("Tx{%X}", []byte(tx))
}

// MerkleHash returns the simple Merkle root hash of the transactions.
func (t *Txs) MerkleHash() []byte {
	txBzs := make([][]byte, len(*t))
	for i := 0; i < len(*t); i++ {
		txBzs[i] = (*t)[i]
	}
	return merkle.SimpleHashFromByteSlices(txBzs)
}

// Index returns the index of this transaction in the list, or -1 if not found
func (t *Txs) Index(tx Tx) int {
	for i := range *t {
		if bytes.Equal((*t)[i], tx) {
			return i
		}
	}
	return -1
}

// IndexByHash returns the index of this transaction hash in the list, or -1 if not found
func (t *Txs) IndexByHash(hash []byte) int {
	for i := range *t {
		if bytes.Equal(t.Hash((*t)[i]), hash) {
			return i
		}
	}
	return -1
}

// Proof returns a simple merkle proof for this node.
// Panics if i < 0 or i >= len(txs)
// TODO: optimize this!
func (t *Txs) Proof(i int) Proof {
	l := len(*t)
	bzs := make([][]byte, l)
	for i := 0; i < l; i++ {
		bzs[i] = (*t)[i]
	}
	root, proofs := merkle.SimpleProofsFromByteSlices(bzs)

	return Proof{
		RootHash: root,
		Data:     (*t)[i],
		Proof:    *proofs[i],
	}
}

// LeafHash returns the hash of the this proof refers to.
func (t *Txs) LeafHash(tx Tx) []byte {
	return t.Hash(tx)
}

// Validate ...
func (t *Txs) Validate(dataHash []byte) error {
	return nil
}

// LeafHash ...
func (p Proof) LeafHash() []byte {
	return herhash.Sum(p.Data)
}

// Validate verifies the proof. It returns nil if the RootHash matches the dataHash argument,
// and if the proof is internally consistent. Otherwise, it returns an error.
func (p Proof) Validate(dataHash []byte) error {
	if !bytes.Equal(dataHash, p.RootHash) {
		return errors.New("Proof matches different data hash")
	}
	if p.Proof.Index < 0 {
		return errors.New("Proof index cannot be negative")
	}
	if p.Proof.Total <= 0 {
		return errors.New("Proof total must be positive")
	}
	valid := p.Proof.Verify(p.RootHash, p.LeafHash())
	if valid != nil {
		return errors.New("Proof is not internally consistent")
	}
	return nil
}
