package account

import (
	"strings"
	"testing"

	"github.com/herdius/herdius-core/accounts/protobuf"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

func TestVerifyAccountBalanceTrue(t *testing.T) {
	accService := NewAccountService()
	account := &protobuf.Account{Balance: 10}
	accService.SetAccount(account)
	accService.SetTxValue(5)
	accService.SetAssetSymbol("HER")
	assert.True(t, accService.VerifyAccountBalance())
}

func TestVerifyAccountBalanceFalse(t *testing.T) {
	account := &protobuf.Account{Balance: 1}
	accService := NewAccountService()
	accService.SetAccount(account)
	accService.SetTxValue(5)
	accService.SetAssetSymbol("HER")
	assert.False(t, accService.VerifyAccountBalance())
}

func TestVerifyExternalAssetBalanceTrue(t *testing.T) {
	eBalance := &protobuf.EBalance{
		Address: "0xD8f647855876549d2623f52126CE40D053a2ef6A",
		Balance: 10,
	}
	eBalances := make(map[string]*protobuf.EBalanceAsset)
	eBalances["ETH"] = &protobuf.EBalanceAsset{}
	eBalances["ETH"].Asset = make(map[string]*protobuf.EBalance)
	eBalances["ETH"].Asset[eBalance.Address] = eBalance
	account := &protobuf.Account{Balance: 10, EBalances: eBalances}
	accService := NewAccountService()
	accService.SetAccount(account)
	accService.SetExtAddress("0xD8f647855876549d2623f52126CE40D053a2ef6A")
	accService.SetTxValue(1)
	accService.SetAssetSymbol("ETH")
	assert.True(t, accService.VerifyAccountBalance())
}

func TestVerifyExternalAssetBalanceFalse(t *testing.T) {
	eBalance := &protobuf.EBalance{
		Address: "0xD8f647855876549d2623f52126CE40D053a2ef6A",
		Balance: 1,
	}
	eBalances := make(map[string]*protobuf.EBalanceAsset)
	eBalances["ETH"] = &protobuf.EBalanceAsset{}
	eBalances["ETH"].Asset = make(map[string]*protobuf.EBalance)
	eBalances["ETH"].Asset[eBalance.Address] = eBalance
	account := &protobuf.Account{Balance: 1, EBalances: eBalances}
	accService := NewAccountService()
	accService.SetAccount(account)
	accService.SetExtAddress("0xD8f647855876549d2623f52126CE40D053a2ef6A")
	accService.SetTxValue(2)
	accService.SetAssetSymbol("ETH")
	assert.False(t, accService.VerifyAccountBalance())
}

func TestVerifyAccountNonceHighTrue(t *testing.T) {
	acc := &protobuf.Account{Nonce: 1}
	txNonce := uint64(3)
	s := NewAccountService()
	res := s.VerifyAccountNonce(acc, txNonce)
	assert.True(t, res)
}

func TestVerifyAccountNonceLowFalse(t *testing.T) {
	acc := &protobuf.Account{Nonce: 1}
	txNonce := uint64(0)
	s := NewAccountService()
	res := s.VerifyAccountNonce(acc, txNonce)
	assert.False(t, res)
}

func TestVerifyAccountNonceTrue(t *testing.T) {
	acc := &protobuf.Account{Nonce: 1}
	txNonce := uint64(2)
	s := NewAccountService()
	res := s.VerifyAccountNonce(acc, txNonce)
	assert.True(t, res)
}

func TestPublicAddressCreation(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	pubKey := privKey.PubKey()

	accService := NewAccountService()
	address, err := accService.GetPublicAddress(pubKey.Bytes())
	assert.NoError(t, err, "Could not create address from public key")

	// Check and verify if herdius address starts with 'H'
	assert.True(t, strings.HasPrefix(address, "H"))
}

func TestIsHerdiusZeroAddressFalse(t *testing.T) {
	accService := NewAccountService()
	accService.SetReceiverAddress("Hx-Not-Zero-Address")
	assert.False(t, accService.IsHerdiusZeroAddress())
}

func TestIsHerdiusZeroAddressTrue(t *testing.T) {
	accService := NewAccountService()
	accService.SetReceiverAddress("Hx00000000000000000000000000000000")
	assert.True(t, accService.IsHerdiusZeroAddress())
}

func TestAccountExternalAddressExistFalse(t *testing.T) {
	accService := NewAccountService()
	eBalance := &protobuf.EBalance{
		Address: "0xD8f647855876549d2623f52126CE40D053a2ef6A",
		Balance: 1,
	}
	eBalances := make(map[string]*protobuf.EBalanceAsset)
	eBalances["ETH"] = &protobuf.EBalanceAsset{}
	eBalances["ETH"].Asset = make(map[string]*protobuf.EBalance)
	eBalances["ETH"].Asset[eBalance.Address] = eBalance
	account := &protobuf.Account{Balance: 1, EBalances: eBalances}
	accService.SetAccount(account)
	accService.SetExtAddress("0xD8f647855876549d2623f52126CE40D053a2ef6A-WrongOne")
	assert.False(t, accService.AccountExternalAddressExist())
}
func TestAccountExternalAddressExistTrue(t *testing.T) {
	accService := NewAccountService()
	eBalance := &protobuf.EBalance{
		Address: "0xD8f647855876549d2623f52126CE40D053a2ef6A",
		Balance: 1,
	}
	eBalances := make(map[string]*protobuf.EBalanceAsset)
	eBalances["ETH"] = &protobuf.EBalanceAsset{}
	eBalances["ETH"].Asset = make(map[string]*protobuf.EBalance)
	eBalances["ETH"].Asset[eBalance.Address] = eBalance
	account := &protobuf.Account{Balance: 1, EBalances: eBalances}
	accService.SetAccount(account)
	accService.SetExtAddress("0xD8f647855876549d2623f52126CE40D053a2ef6A")
	assert.False(t, accService.AccountExternalAddressExist())
}
