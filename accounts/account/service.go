package account

import (
	"fmt"
	"log"
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
	herdiusZeroAddress      = "Hx00000000000000000000000000000000"
	numAddressPerAssetLimit = 1024
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
type Service struct {
	account         *protobuf.Account
	assetSymbol     string
	address         string
	receiverAddress string
	extAddress      string
	txValue         uint64
	txLockedAmount  uint64
	txRedeemAmount  uint64
}

// Account returns state db account
func (s *Service) Account() *protobuf.Account {
	return s.account
}

// SetAccount sets Service account
func (s *Service) SetAccount(account *protobuf.Account) {
	s.account = account
}

// AssetSymbol returns servie asset symbol
func (s *Service) AssetSymbol() string {
	return s.assetSymbol
}

// SetAssetSymbol sets Service asset symbol
func (s *Service) SetAssetSymbol(assetSymbol string) {
	s.assetSymbol = assetSymbol
}

// Address returns service her account address
func (s *Service) Address() string {
	return s.address
}

// SetAddress sets Service asset symbol
func (s *Service) SetAddress(address string) {
	s.address = address
}

// ReceiverAddress returns receiver's her account address
func (s *Service) ReceiverAddress() string {
	return s.receiverAddress
}

// SetReceiverAddress sets receiver's her account address
func (s *Service) SetReceiverAddress(receiverAddress string) {
	s.receiverAddress = receiverAddress
}

// ExtAddress returns service ExtAddress
func (s *Service) ExtAddress() string {
	return s.extAddress
}

// SetExtAddress sets Service ExtAddress
func (s *Service) SetExtAddress(extAddress string) {
	s.extAddress = extAddress
}

// TxValue returns transaction transfer value
func (s *Service) TxValue() uint64 {
	return s.txValue
}

// SetTxValue sets transaction transfer value
func (s *Service) SetTxValue(txValue uint64) {
	s.txValue = txValue
}

// TxLockedAmount returns transaction transfer locked amount
func (s *Service) TxLockedAmount() uint64 {
	return s.txLockedAmount
}

// SetTxLockedAmount sets transaction transfer locked amount
func (s *Service) SetTxLockedAmount(txLockedAmount uint64) {
	s.txLockedAmount = txLockedAmount
}

// TxRedeemAmount returns transaction transfer redeem amount
func (s *Service) TxRedeemAmount() uint64 {
	return s.txRedeemAmount
}

// SetTxRedeemAmount sets transaction transfer redeem amount
func (s *Service) SetTxRedeemAmount(txRedeemAmount uint64) {
	s.txRedeemAmount = txRedeemAmount
}

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

	lockBalances := make(map[string]*protobuf.LockBalanceAsset)

	for asset, assetAccount := range account.LockedBalance {
		lockBalances[asset] = &protobuf.LockBalanceAsset{}
		lockBalances[asset].Asset = make(map[string]uint64)
		for addr, lockAmount := range assetAccount {
			lockBalances[asset].Asset[addr] = lockAmount
		}
	}

	acc := &protobuf.Account{
		PublicKey:            account.PublicKey,
		Address:              account.Address,
		Nonce:                account.Nonce,
		Balance:              account.Balance,
		StorageRoot:          stateRootHex.String(),
		EBalances:            eBalances,
		LockBalances:         lockBalances,
		Erc20Address:         account.Erc20Address,
		ExternalNonce:        account.ExternalNonce,
		LastBlockHeight:      account.LastBlockHeight,
		FirstExternalAddress: account.FirstExternalAddress,
	}
	return acc, nil
}

// VerifyAccountBalance verifies if account has enough HER tokens or external asset balances
func (s *Service) VerifyAccountBalance() bool {
	symbol := strings.ToUpper(s.assetSymbol)
	// Get the balance of required asset
	if strings.EqualFold(symbol, "HER") {
		if s.account.Balance >= s.txValue {
			return true
		}
	} else if s.account != nil && len(s.account.EBalances) > 0 && s.account.EBalances[symbol] != nil {
		lockedAmount := uint64(0)
		if s.account.LockBalances[symbol] != nil {
			lockedAmount = s.account.LockBalances[symbol].Asset[s.extAddress]
		}
		if asset := s.account.EBalances[symbol].Asset; asset != nil {
			eb, ok := asset[s.extAddress]
			if ok && eb.Balance > lockedAmount {
				return (eb.Balance - lockedAmount) >= s.txValue
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
func (s *Service) AccountExternalAddressExist() bool {
	if s.account != nil && s.account.EBalances != nil && s.account.EBalances[s.assetSymbol] != nil {
		if asset := s.account.EBalances[s.assetSymbol].Asset; asset != nil {
			_, ok := asset[s.extAddress]
			return ok
		}
	}
	return false
}

// IsHerdiusZeroAddress checks if receiver address is of herdius zero address
// when lock tx type is transmitted to herdius blockchain
func (s *Service) IsHerdiusZeroAddress() bool {
	return s.receiverAddress == herdiusZeroAddress
}

// AccountEBalancePerAssetReachLimit reports whether an account reaches limit for number of address per asset.
func (s *Service) AccountEBalancePerAssetReachLimit() bool {
	if s.account != nil && s.account.EBalances != nil && s.account.EBalances[s.assetSymbol] != nil {
		return len(s.account.EBalances[s.assetSymbol].Asset) >= numAddressPerAssetLimit
	}
	return false
}

// VerifyLockedAmount checks account have enough external balance for lock.
func (s *Service) VerifyLockedAmount() bool {
	if s.account != nil && s.account.EBalances != nil && s.account.EBalances[s.assetSymbol] != nil {
		if asset := s.account.EBalances[s.assetSymbol].Asset; asset != nil {
			return s.txLockedAmount <= asset[s.extAddress].Balance
		}
	}
	return false
}

// VerifyRedeemAmount checks account have proper locked amount for redeeming
func (s *Service) VerifyRedeemAmount() bool {
	log.Printf("Account before Redeem: %+v", s.account)
	if s.account != nil && s.account.LockBalances != nil && s.account.LockBalances[s.assetSymbol] != nil {
		if asset := s.account.LockBalances[s.assetSymbol].Asset; asset != nil {
			return s.txRedeemAmount <= asset[s.extAddress]
		}
	}
	return false
}
