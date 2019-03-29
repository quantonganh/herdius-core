package account

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/herdius/herdius-core/accounts/protobuf"
	"github.com/herdius/herdius-core/blockchain"
	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/storage/state/statedb"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {

	cryptoAmino.RegisterAmino(cdc)
}

// ServiceI is an account service interface
type ServiceI interface {
	GetAccountByAddress(address string) (*protobuf.Account, error)
}

// Service ...
type Service struct{}

// TODO: Make it better so as not a
// global variable
var (
	_ ServiceI = (*Service)(nil)
)

func NewAccountService() *Service {
	return &Service{}
}

func (s *Service) GetPublicAddress(keyBytes []byte) (string, error) {
	pubkey := secp256k1.PubKeySecp256k1{}
	cdc.UnmarshalBinaryBare(keyBytes, &pubkey)
	herAddress := pubkey.GetAddress()

	return herAddress, nil
}

func (s *Service) GetAccountByAddress(address string) (*protobuf.Account, error) {
	blockchainSvc := &blockchain.Service{}
	lastBlock := blockchainSvc.GetLastBlock()

	var stateRootHex cmn.HexBytes
	stateRoot := lastBlock.GetHeader().GetStateRoot()
	stateRootHex = stateRoot

	// Get Trie Root of state db from last block
	stateTrie, err := ethtrie.New(common.BytesToHash(stateRoot), statedb.GetDB())

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to retrieve the state trie: %v.", err))
	}

	pubKeyBytes := []byte(address)
	actbz, err := stateTrie.TryGet(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to retrieve account detail of address:%s. %v.", address, err))
	}
	var account statedb.Account
	err = cdc.UnmarshalJSON(actbz, &account)

	if len(account.Address) == 0 {
		return nil, nil
	}

	acc := &protobuf.Account{
		Address:     account.Address,
		Nonce:       account.Nonce,
		Balance:     account.Balance,
		StorageRoot: stateRootHex.String(),
	}
	return acc, nil
}
