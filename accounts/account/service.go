package account

import (
	"fmt"
	"strings"

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
	eBalances := make(map[string]*protobuf.EBalanceAsset)

	for asset, assetAccount := range account.EBalances {
		eBalances[asset] = &protobuf.EBalanceAsset{}
		eBalances[asset].Asset = make(map[string]*protobuf.EBalance)
		for _, eb := range assetAccount {
			eBalanceRes := &protobuf.EBalance{
				Address:         eb.Address,
				Balance:         eb.Balance,
				LastBlockHeight: eb.LastBlockHeight,
				Nonce:           eb.Nonce,
			}
			eBalances[asset].Asset[eb.Address] = eBalanceRes
		}
	}

	acc := &protobuf.Account{
		PublicKey:       account.PublicKey,
		Address:         account.Address,
		Nonce:           account.Nonce,
		Balance:         account.Balance,
		StorageRoot:     stateRootHex.String(),
		EBalances:       eBalances,
		Erc20Address:    account.Erc20Address,
		ExternalNonce:   account.ExternalNonce,
		LastBlockHeight: account.LastBlockHeight,
	}
	return acc, nil
}

// VerifyAccountBalance verifies if account has enough HER tokens
// This only has to be verified and called for HER crypto asset
func (s *Service) VerifyAccountBalance(a *protobuf.Account, txValue uint64, assetSymbol string) bool {
	// Get the balance of required asset
	if strings.EqualFold(assetSymbol, "HER") {
		if a.Balance >= txValue {
			return true
		}
	} else if a != nil && len(a.EBalances) > 0 && a.EBalances[assetSymbol] != nil {
		for _, eb := range a.EBalances[assetSymbol].Asset {
			if eb.Balance >= txValue {
				return true
			}
		}
	}
	return false
}

// VerifyAccountNonce verifies initiated transaction has Nonce value greater than
// Nonce value in account
func (s *Service) VerifyAccountNonce(a *protobuf.Account, txNonce uint64) bool {
	if a.Nonce == 0 && txNonce == 1 {
		return true
	}
	if txNonce >= a.Nonce+1 {
		return true
	}
	return false
}
