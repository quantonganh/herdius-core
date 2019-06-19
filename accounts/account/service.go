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

const (
	herdiusZeroAddress = "Hx00000000000000000000000000000000"
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

// VerifyAccountBalance verifies if account has enough HER tokens or external asset balances
func (s *Service) VerifyAccountBalance(a *protobuf.Account, txValue uint64, assetSymbol string, extAddress string) bool {
	// Get the balance of required asset
	if strings.EqualFold(strings.ToUpper(assetSymbol), "HER") {
		if a.Balance >= txValue {
			return true
		}
	} else if a != nil && len(a.EBalances) > 0 && a.EBalances[strings.ToUpper(assetSymbol)] != nil {
		if asset := a.EBalances[strings.ToUpper(assetSymbol)].Asset; asset != nil {
			eb, ok := asset[extAddress]
			if ok && eb.Balance >= txValue {
				return ok
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

// AccountExternalAddressExist reports whether an external address existed in EBalances
func (s *Service) AccountExternalAddressExist(a *protobuf.Account, assetSymbol, extAddress string) bool {
	if a != nil && a.EBalances != nil && a.EBalances[assetSymbol] != nil {
		if asset := a.EBalances[assetSymbol].Asset; asset != nil {
			_, ok := asset[extAddress]
			return ok
		}
	}
	return false
}

// IsHerdiusZeroAddress checks if receiver address is of herdius zero address
// when lock tx type is transmitted to herdius blockchain
func (s *Service) IsHerdiusZeroAddress(address string) bool {
	if address == herdiusZeroAddress {
		return true
	}
	return false
}
