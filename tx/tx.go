package tx

import (
	"github.com/herdius/herdius-core/crypto/merkle"
	cmn "github.com/herdius/herdius-core/libs/common"
)

// Tx is an arbitrary byte array.
type Tx []byte

// Txs is a slice of Tx.
type Txs [][]byte

// Proof represents a Merkle proof of the presence of a transaction in the Merkle tree.
type Proof struct {
	RootHash cmn.HexBytes
	Data     Tx
	Proof    merkle.SimpleProof
}
