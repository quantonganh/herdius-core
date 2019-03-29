package statedb

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"

	tt "github.com/ethereum/go-ethereum/trie"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

func newTestLDB() (*ethdb.LDBDatabase, func()) {
	dirname, err := ioutil.TempDir(os.TempDir(), "ethdb_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := ethdb.NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return db, func() {
		db.Close()
		os.RemoveAll(dirname)
	}
}

func TestTryIt(t *testing.T) {
	db, remove := newTestLDB()
	defer remove()
	k := "key"
	err := db.Put([]byte(k), nil)
	assert.NoError(t, err, fmt.Sprintf("can't add (key, value) to Trie: %v", err))
}
func TestCreateStateSingleton(t *testing.T) {
	dir, err := ioutil.TempDir("", "trie-singleton")

	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory: %v", err))
	}

	assert.NoError(t, err)

	trie := GetState(dir)

	assert.NotNil(t, trie)

	assert.Equal(t, trie, GetState(dir))

	os.RemoveAll(dir)
}

func TestTryGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "trie-singleton")

	assert.NoError(t, err, fmt.Sprintf("can't create temporary directory: %v", err))

	trie := GetState(dir)
	err = trie.TryUpdate([]byte("key"), []byte("value"))
	assert.NoError(t, err, fmt.Sprintf("can't add (key, value) to Trie: %v", err))

	root, err := trie.Commit(nil)
	assert.NoError(t, err, fmt.Sprintf("can't commit (key, value) to Trie: %v", err))

	assert.NotNil(t, root)

	v, err := trie.TryGet(([]byte("key"))[:])
	assert.NoError(t, err, fmt.Sprintf("can't get value from Trie: %v", err))
	value := string(v)
	assert.Equal(t, "value", value)
	os.RemoveAll(dir)
}

// func TestTryUpdateGetAccounts(t *testing.T) {
// 	var accountList []Account

// 	dir, err := ioutil.TempDir("", "trie-singleton")

// 	assert.NoError(t, err, fmt.Sprintf("can't create temporary directory: %v", err))

// 	trie := GetState(dir)

// 	for i := 0; i < 10; i++ {
// 		account := createAccount(i + 1)
// 		js, err := cdc.MarshalJSON(account)
// 		assert.NoError(t, err, fmt.Sprintf("can't encode the account object: %v", err))
// 		accountList = append(accountList, account)

// 		err = trie.TryUpdate(account.AddressHash, js)

// 		assert.NoError(t, err, fmt.Sprintf("can't add (key, value) to Trie: %v", err))
// 	}

// 	root, err := trie.Commit(nil)
// 	assert.NoError(t, err, fmt.Sprintf("can't commit (key, value) to Trie: %v", err))

// 	assert.NotNil(t, root)

// 	//////////////////
// 	// Retrieve and Verify the Accounts using root hash
// 	trie2, err := tt.New(common.BytesToHash(root), GetDB())
// 	assert.NoError(t, err, fmt.Sprintf("unable to get Trie: %v", err))
// 	assert.NotNil(t, trie2)

// 	size := len(accountList)

// 	for i := 0; i < size; i++ {
// 		account := accountList[i]
// 		a, _ := trie2.TryGet((account.AddressHash))
// 		var av Account
// 		err = cdc.UnmarshalJSON(a, &av)

// 		assert.NoError(t, err, fmt.Sprintf("unable to unmarshal: %v", err))

// 		assert.NotNil(t, av)
// 		assert.Equal(t, uint64(i+1), av.Nonce)

// 	}

// 	os.RemoveAll(dir)
// }

func TestTryGetInDir(t *testing.T) {
	dir := "./herdius/statedb"
	trie := GetState(dir)
	err := trie.TryUpdate([]byte("key"), []byte("value"))
	assert.NoError(t, err, fmt.Sprintf("can't add (key, value) to Trie: %v", err))

	root, err := trie.Commit(nil)
	assert.NoError(t, err, fmt.Sprintf("can't commit (key, value) to Trie: %v", err))

	assert.NotNil(t, root)

	v, err := trie.TryGet(([]byte("key"))[:])
	assert.NoError(t, err, fmt.Sprintf("can't get value from Trie: %v", err))
	value := string(v)
	assert.Equal(t, "value", value)
	os.RemoveAll(dir)
}

// func createAccount(nonce int) Account {
// 	var privKey = ed25519.GenPrivKey()
// 	var pubKey = privKey.PubKey()

// 	address := fmt.Sprintf("%X", pubKey.Address().Bytes()[:])
// 	var bal = "100"
// 	account := Account{
// 		Nonce:       uint64(nonce),
// 		Address:     address,
// 		AddressHash: pubKey.Bytes(),
// 		Balance:     []byte(bal),
// 	}

// 	return account
// }

func TestTryUpdateGetSecp256k1Accounts(t *testing.T) {
	var accountList []Account

	dir, err := ioutil.TempDir("", "trie-singleton")

	assert.NoError(t, err, fmt.Sprintf("can't create temporary directory: %v", err))

	trie := GetState(dir)

	for i := 0; i < 10; i++ {
		account := createSecp256k1Account(i + 1)
		js, err := cdc.MarshalJSON(account)
		assert.NoError(t, err, fmt.Sprintf("can't encode the account object: %v", err))
		accountList = append(accountList, account)

		err = trie.TryUpdate(account.AddressHash, js)

		assert.NoError(t, err, fmt.Sprintf("can't add (key, value) to Trie: %v", err))
	}

	root, err := trie.Commit(nil)
	assert.NoError(t, err, fmt.Sprintf("can't commit (key, value) to Trie: %v", err))

	assert.NotNil(t, root)

	// Retrieve and Verify the Accounts using root hash
	trie2, err := tt.New(common.BytesToHash(root), GetDB())
	assert.NoError(t, err, fmt.Sprintf("unable to get Trie: %v", err))
	assert.NotNil(t, trie2)

	size := len(accountList)

	for i := 0; i < size; i++ {
		account := accountList[i]
		a, _ := trie2.TryGet((account.AddressHash))
		var av Account
		err = cdc.UnmarshalJSON(a, &av)

		assert.NoError(t, err, fmt.Sprintf("unable to unmarshal: %v", err))

		assert.NotNil(t, av)
		assert.Equal(t, uint64(i+1), av.Nonce)

	}

	os.RemoveAll(dir)
}

func createSecp256k1Account(nonce int) Account {
	var privKey = secp256k1.GenPrivKey()
	var pubKey = privKey.PubKey()

	address := fmt.Sprintf("%X", pubKey.Address().Bytes()[:])

	account := Account{
		Nonce:       uint64(nonce),
		Address:     address,
		AddressHash: pubKey.Bytes(),
		Balance:     100,
	}

	return account
}
