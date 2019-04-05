package account

import (
	"errors"
	"fmt"

	b64 "encoding/base64"

	"github.com/ethereum/go-ethereum/common"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/herdius/herdius-core/accounts/protobuf"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/crypto"
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
	RegisterAccount(request *protobuf.AccountRegisterRequest) (*protobuf.Account, error)
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
		PublicKey:   account.PublicKey,
		Address:     account.Address,
		Nonce:       account.Nonce,
		Balance:     account.Balance,
		StorageRoot: stateRootHex.String(),
	}
	return acc, nil
}

// RegisterAccount registers a new account to state trie database
// TODO: This registeration process should only be used when
//		 one supervisor node is running within Herdius Network.
//		 It will need a transaction to be sent to herdius blockchain
// 		 as well in case multiple supervisor nodes are running so that
//		 every supervisor node will have the same account state.
func (s *Service) RegisterAccount(request *protobuf.AccountRegisterRequest) (*protobuf.Account, error) {
	if request != nil && len(request.PublicKey) > 0 {
		return nil, errors.New("Mising Public key")
	}

	accountBz, err := b64.StdEncoding.DecodeString(request.PublicKey)
	if err != nil {
		return nil, errors.New("Failed to register account due to : " + err.Error())
	}
	var pubKey crypto.PubKey
	err = cdc.UnmarshalBinaryBare(accountBz, &pubKey)
	if err != nil {
		return nil, errors.New("Failed to register account due to : " + err.Error())
	}
	blockchainSvc := &blockchain.Service{}
	lastBlock := blockchainSvc.GetLastBlock()

	//var stateRootHex cmn.HexBytes
	stateRoot := lastBlock.GetHeader().GetStateRoot()
	//stateRootHex = stateRoot

	// Get Trie Root of state db from last block
	stateTrie, err := ethtrie.New(common.BytesToHash(stateRoot), statedb.GetDB())

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to retrieve the state trie: %v.", err))
	}
	// Check if account is already registered
	address := pubKey.GetAddress()
	if len(address) > 0 {

		pubKeyBytes := []byte(address)
		actbz, err := stateTrie.TryGet(pubKeyBytes)
		if err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("Failed to retrieve account detail of address:%s. %v.", address, err))
		}

		var account statedb.Account
		err = cdc.UnmarshalJSON(actbz, &account)

		if len(account.Address) > 0 {
			return &protobuf.Account{
				PublicKey:   account.PublicKey,
				Address:     account.Address,
				Balance:     account.Balance,
				Nonce:       account.Nonce,
				StorageRoot: account.StateRoot,
			}, nil
		}
	}

	// Register a new account
	account := statedb.Account{
		PublicKey: request.PublicKey,
		Nonce:     0,
		Address:   address,
		Balance:   0,
	}

	actbz, _ := cdc.MarshalJSON(account)
	addressBz := []byte(address)

	err = stateTrie.TryUpdate(addressBz, actbz)
	if err != nil {
		return nil, errors.New("Failed to register account due to : " + err.Error())
	}

	root, err := stateTrie.Commit(nil)
	lastBlock.Header.StateRoot = root.Bytes()
	err = blockchainSvc.AddBaseBlock(lastBlock)
	if err != nil {
		return nil, errors.New("Failed to register account due to : " + err.Error())
	}
	return &protobuf.Account{
		PublicKey:   account.PublicKey,
		Address:     account.Address,
		Balance:     account.Balance,
		Nonce:       account.Nonce,
		StorageRoot: b64.StdEncoding.EncodeToString(lastBlock.Header.StateRoot),
	}, nil
}

// VerifyAccountBalance verifies if account has enough HER tokens
// This only has to be verified and called for HER crypto asset
func (s *Service) VerifyAccountBalance(a *protobuf.Account, txValue uint64) bool {
	if a.Balance < txValue {
		return false
	}
	return true
}

// VerifyAccountNonce verifies initiated transaction has Nonce value greater than
// Nonce value in account
func (s *Service) VerifyAccountNonce(a *protobuf.Account, txNonce uint64) bool {
	if txNonce > a.Nonce {
		return true
	}
	return false
}
