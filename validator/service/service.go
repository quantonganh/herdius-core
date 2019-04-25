package service

import (
	"fmt"

	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/crypto"
	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/crypto/herhash"
	"github.com/herdius/herdius-core/crypto/merkle"
	"github.com/herdius/herdius-core/p2p/network"
	"github.com/herdius/herdius-core/supervisor/transaction"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	RegisterValidatorServiceAmino(cdc)
}

// RegisterValidatorServiceAmino ...
func RegisterValidatorServiceAmino(cdc *amino.Codec) {
	cryptoAmino.RegisterAmino(cdc)

}

// ValidatorI is an interface for Validators
type ValidatorI interface {
	VerifyTxs(rootHash []byte, txs [][]byte) error
	Vote(net *network.Network, address string, cbm *protobuf.ChildBlockMessage) error
}

// Validator concrete implementation of ValidatorI
type Validator struct{}

var (
	_ ValidatorI = (*Validator)(nil)
)

// VerifyTxs verifies the merkel root hash of the Txs
func (v *Validator) VerifyTxs(rootHash []byte, txs [][]byte) error {
	rootHash2, proofs := merkle.SimpleProofsFromByteSlices(txs)
	// # of Txs in each batch
	total := 500

	if rootHash2 == nil || proofs == nil {
		return fmt.Errorf(fmt.Sprintf("Unmatched root hashes: %X vs %X", rootHash, rootHash2))
	}
	for i, tx := range txs {
		txHash := herhash.Sum(tx)
		proof := proofs[i]

		if proof.Index != i {
			return fmt.Errorf(fmt.Sprintf("Unmatched indicies: %d vs %d", proof.Index, i))
		}

		// TODO: pass total number of transactions in the batch
		// Right now it is 500 txs
		if proof.Total != total {
			return fmt.Errorf(fmt.Sprintf("Unmatched totals: %d vs %d", proof.Total, total))
		}

		// Verify success
		err := proof.Verify(rootHash, txHash)

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Proof Verification failed: %v.", err))
		}

		//Verify TX signature
		txValue := transaction.Tx{}
		err = cdc.UnmarshalJSON(tx, &txValue)

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("TX Unmarshaling failed: %v.", err))
		}

		msg := txValue.Message
		var pubkey crypto.PubKey
		err = cdc.UnmarshalBinaryBare([]byte(txValue.SenderPubKey), &pubkey)

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Pub Key Unmarshaling failed: %v.", err))
		}

		isVerified := pubkey.VerifyBytes([]byte(msg), []byte(txValue.Signature))
		if !isVerified {
			return fmt.Errorf(fmt.Sprintf("TX signature verification failed: %v.", isVerified))
		}
	}
	return nil
}

// Vote adds validator details and it's sign
func (v *Validator) Vote(net *network.Network, address string, cbm *protobuf.ChildBlockMessage) error {

	keys := net.GetKeys()

	// TODO: Staking power needs to be checked and updated from the state db and updated
	validator := &protobuf.Validator{
		Address:      address,
		PubKey:       keys.PubKey.Bytes(),
		Stakingpower: 100,
	}

	validator.PubKey = keys.PubKey.Bytes()

	//TODO : It is signed by public key of the validator
	// It needs to be changed to something useful
	sign, err := keys.PrivKey.Sign(validator.PubKey)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to sign the vote: %v.", sign))
	}
	vote := &protobuf.VoteInfo{
		Validator:          validator,
		Signature:          sign,
		SignedCurrentBlock: true,
	}

	cbm.Vote = vote

	return nil
}
