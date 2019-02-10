package tx

import (
	"bytes"
	"errors"
	"fmt"

	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/crypto/herhash"
	"github.com/herdius/herdius-core/crypto/merkle"
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
	Hash(tx Tx) []byte           // Hash computes the hash of the wire encoded transaction.
	String(tx Tx) string         // String returns the hex-encoded transaction as a string.
	Index(tx Tx) int             // Index returns the index of this transaction in the list, or -1 if not found
	IndexByHash(hash []byte) int // IndexByHash returns the index of this transaction hash in the list, or -1 if not found
	MerkleHash() []byte          // MerkleHash returns the simple Merkle root hash of the transactions.
	Proof(i int) Proof           // Proof returns a simple merkle proof for this node.
	LeafHash(tx Tx) []byte       // LeafHash returns the hash of the this proof refers to.
	// Validate verifies the proof. It returns nil if the RootHash matches the dataHash argument,
	// and if the proof is internally consistent. Otherwise, it returns a sensible error.
	Validate(dataHash []byte) error
}

type tx struct {
	txs [][]byte
}

// GetTxService creates a new transaction service
func GetTxService() Service {
	return &tx{}
}
func (t *tx) SetTxs(txs [][]byte) {
	t.txs = txs
}

// Hash ...
func (t *tx) Hash(tx Tx) []byte {
	return herhash.Sum(tx)
}

// String returns the hex-encoded transaction as a string.
func (t *tx) String(tx Tx) string {
	return fmt.Sprintf("Tx{%X}", []byte(tx))
}

// MerkleHash returns the simple Merkle root hash of the transactions.
func (t *tx) MerkleHash() []byte {
	txBzs := make([][]byte, len(t.txs))
	for i := 0; i < len(t.txs); i++ {
		txBzs[i] = t.txs[i]
	}
	return merkle.SimpleHashFromByteSlices(txBzs)
}

// Index returns the index of this transaction in the list, or -1 if not found
func (t *tx) Index(tx Tx) int {
	for i := range t.txs {
		if bytes.Equal(t.txs[i], tx) {
			return i
		}
	}
	return -1
}

// IndexByHash returns the index of this transaction hash in the list, or -1 if not found
func (t *tx) IndexByHash(hash []byte) int {
	for i := range t.txs {
		if bytes.Equal(t.Hash(t.txs[i]), hash) {
			return i
		}
	}
	return -1
}

// Proof returns a simple merkle proof for this node.
// Panics if i < 0 or i >= len(txs)
// TODO: optimize this!
func (t *tx) Proof(i int) Proof {
	l := len(t.txs)
	bzs := make([][]byte, l)
	for i := 0; i < l; i++ {
		bzs[i] = t.txs[i]
	}
	root, proofs := merkle.SimpleProofsFromByteSlices(bzs)

	return Proof{
		RootHash: root,
		Data:     t.txs[i],
		Proof:    *proofs[i],
	}
}

// LeafHash returns the hash of the this proof refers to.
func (t *tx) LeafHash(tx Tx) []byte {
	return t.Hash(tx)
}

// Validate ...
func (t *tx) Validate(dataHash []byte) error {
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
