package account

import (
	"strings"
	"testing"

	"github.com/herdius/herdius-core/accounts/protobuf"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

func TestVerifyAccountBalanceTrue(t *testing.T) {
	balances := make(map[string]uint64)
	balances["HER"] = 10

	account := &protobuf.Account{Balance: 10, Balances: balances}

	accService := NewAccountService()
	assert.True(t, accService.VerifyAccountBalance(account, 5, "HER"))
}

func TestVerifyAccountBalanceFalse(t *testing.T) {
	balances := make(map[string]uint64)
	balances["HER"] = 1

	account := &protobuf.Account{Balances: balances}

	accService := NewAccountService()
	assert.False(t, accService.VerifyAccountBalance(account, 5, "HER"))
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
