package account

import (
	"strings"
	"testing"

	"github.com/herdius/herdius-core/accounts/protobuf"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

func TestVerifyAccountBalanceTrue(t *testing.T) {
	account := &protobuf.Account{Balance: 10}
	accService := NewAccountService()
	assert.True(t, accService.VerifyAccountBalance(account, 5, "HER", ""))
}

func TestVerifyAccountBalanceFalse(t *testing.T) {
	account := &protobuf.Account{Balance: 1}
	accService := NewAccountService()
	assert.False(t, accService.VerifyAccountBalance(account, 5, "HER", ""))
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
	assert.True(t, accService.VerifyAccountBalance(account, 5, "ETH", eBalance.Address))
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
	assert.False(t, accService.VerifyAccountBalance(account, 5, "ETH", eBalance.Address))
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
