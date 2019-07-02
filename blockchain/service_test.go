package blockchain

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	pluginproto "github.com/herdius/herdius-core/hbi/protobuf"
	"github.com/herdius/herdius-core/storage/db"

	"github.com/herdius/herdius-core/crypto/herhash"
	txbyte "github.com/herdius/herdius-core/tx"
	"github.com/stretchr/testify/require"
)

// Adds 4 Blocks to blockchain db. Each block will consist of 5 transactions
// This is performed for a single account address.
// Total Number of Transactions = 20
func TestGetTxsByAddress(t *testing.T) {
	dirname, err := ioutil.TempDir(os.TempDir(), "badgerdb_test_")
	LoadDBTest(dirname)
	defer os.RemoveAll(dirname)
	defer badgerDB.Close() // Close the db to release the lock
	require.Nil(t, err)

	privKey := secp256k1.GenPrivKey()
	addBlocks(privKey, t)
	txSrv := TxService{}
	txs, err := txSrv.GetTxs(privKey.PubKey().GetAddress())
	require.Nil(t, err)

	assert.Equal(t, 20, len(txs.GetTxs()), "Total fetched transactions should be 20")
}

func TestGetTxsByAddressAndAsset(t *testing.T) {
	dirname, err := ioutil.TempDir(os.TempDir(), "badgerdb_test_")
	LoadDBTest(dirname)
	defer os.RemoveAll(dirname)
	defer badgerDB.Close() // Close the db to release the lock
	require.Nil(t, err)

	privKey := secp256k1.GenPrivKey()
	addBlocks(privKey, t)
	txSrv := TxService{}
	txs, err := txSrv.GetTxsByAssetAndAddress("HER", privKey.PubKey().GetAddress())
	require.Nil(t, err)

	assert.Equal(t, 1, len(txs.GetTxs()), "Total HER transactions should be 1")

	btcTxs, err := txSrv.GetTxsByAssetAndAddress("BTC", privKey.PubKey().GetAddress())
	require.Nil(t, err)

	assert.Equal(t, 0, len(btcTxs.GetTxs()), "Total BTC transactions should be 0")
}

func addBlocks(privKey secp256k1.PrivKeySecp256k1, t *testing.T) {
	var txsBatch1 txbyte.Txs
	var txsBatch2 txbyte.Txs
	var txsBatch3 txbyte.Txs
	var txsBatch4 txbyte.Txs

	for i := 1; i <= 5; i++ {
		tx := getTx(i, privKey)
		txbz, err := cdc.MarshalJSON(tx)
		require.Nil(t, err)
		txsBatch1 = append(txsBatch1, txbz)
	}
	assert.Equal(t, 5, len(txsBatch1))
	// Block 1
	baseBlock1 := createBlock(1, txsBatch1, t)
	blockhash := baseBlock1.GetHeader().GetBlock_ID().GetBlockHash()
	bbbz1, err := cdc.MarshalJSON(baseBlock1)
	require.Nil(t, err)
	badgerDB.Set(blockhash, bbbz1)

	for i := 6; i <= 10; i++ {
		tx := getTx(i, privKey)
		txbz, err := cdc.MarshalJSON(tx)
		require.Nil(t, err)
		txsBatch2 = append(txsBatch2, txbz)
	}
	assert.Equal(t, 5, len(txsBatch2))
	// Block 2
	baseBlock2 := createBlock(2, txsBatch2, t)
	blockhash = baseBlock2.GetHeader().GetBlock_ID().GetBlockHash()
	bbbz2, err := cdc.MarshalJSON(baseBlock2)
	require.Nil(t, err)
	badgerDB.Set(blockhash, bbbz2)

	for i := 11; i <= 15; i++ {
		tx := getTx(i, privKey)
		txbz, err := cdc.MarshalJSON(tx)
		require.Nil(t, err)
		txsBatch3 = append(txsBatch3, txbz)
	}
	assert.Equal(t, 5, len(txsBatch3))
	// Block 3
	baseBlock3 := createBlock(3, txsBatch3, t)
	blockhash = baseBlock3.GetHeader().GetBlock_ID().GetBlockHash()
	bbbz3, err := cdc.MarshalJSON(baseBlock3)
	require.Nil(t, err)
	badgerDB.Set(blockhash, bbbz3)

	for i := 16; i <= 20; i++ {
		tx := getTx(i, privKey)
		txbz, err := cdc.MarshalJSON(tx)
		require.Nil(t, err)
		txsBatch4 = append(txsBatch4, txbz)
	}
	assert.Equal(t, 5, len(txsBatch4))
	// Block 4
	baseBlock4 := createBlock(4, txsBatch4, t)
	blockhash = baseBlock4.GetHeader().GetBlock_ID().GetBlockHash()
	bbbz4, err := cdc.MarshalJSON(baseBlock4)
	require.Nil(t, err)
	badgerDB.Set(blockhash, bbbz4)
}

func getTx(nonce int, privKey secp256k1.PrivKeySecp256k1) pluginproto.Tx {
	message := "Transfer 10 HER"
	msg := []byte(message)
	pubKey := privKey.PubKey()
	sign, _ := privKey.Sign(msg)
	asset := &pluginproto.Asset{
		Symbol: "HER",
	}
	tx := pluginproto.Tx{
		SenderAddress: pubKey.GetAddress(),
		Message:       message,
		Sign:          string(sign),
		Asset:         asset,
	}
	return tx
}

func createBlock(height int64, txs txbyte.Txs, t *testing.T) *protobuf.BaseBlock {
	ts := time.Now().UTC()
	baseHeader := protobuf.BaseHeader{
		Block_ID: &protobuf.BlockID{},
		Height:   height,
		Time: &protobuf.Timestamp{
			Seconds: ts.Unix(),
			Nanos:   ts.UnixNano(),
		},
		TotalTxs: uint64(len(txs)),
	}

	blockHashBz, err := cdc.MarshalJSON(baseHeader)
	require.Nil(t, err)

	blockHash := herhash.Sum(blockHashBz)
	baseHeader.GetBlock_ID().BlockHash = blockHash
	baseBlock := &protobuf.BaseBlock{
		Header:  &baseHeader,
		TxsData: &protobuf.TxsData{Tx: txs},
	}
	return baseBlock
}

func LoadDBTest(dirname string) {
	badgerDB = db.NewDB("badger", db.GoBadgerBackend, dirname)
}

func addBlocksWithLockedTxs(privKey secp256k1.PrivKeySecp256k1, t *testing.T) int64 {
	var txsBatch txbyte.Txs

	for i := 1; i <= 5; i++ {
		tx := getTx(i, privKey)
		txbz, err := cdc.MarshalJSON(tx)
		require.Nil(t, err)
		txsBatch = append(txsBatch, txbz)
	}
	for i := 1; i <= 5; i++ {
		tx := getTx(i, privKey)
		tx.Type = "Lock"
		txbz, err := cdc.MarshalJSON(tx)
		require.Nil(t, err)
		txsBatch = append(txsBatch, txbz)
	}
	assert.Equal(t, 10, len(txsBatch))

	baseBlock := createBlock(1, txsBatch, t)
	blockhash := baseBlock.GetHeader().GetBlock_ID().GetBlockHash()
	bbbz, err := cdc.MarshalJSON(baseBlock)
	require.Nil(t, err)
	badgerDB.Set(blockhash, bbbz)
	return baseBlock.GetHeader().GetHeight()
}

func TestGetLockedTxsByBlockNumber(t *testing.T) {
	dirname, err := ioutil.TempDir(os.TempDir(), "badgerdb_test_")
	require.Nil(t, err)
	LoadDBTest(dirname)
	defer os.RemoveAll(dirname)
	defer badgerDB.Close() // Close the db to release the lock

	privKey := secp256k1.GenPrivKey()
	blockNumber := addBlocksWithLockedTxs(privKey, t)
	txSrv := TxService{}
	txs, err := txSrv.GetLockedTxsByBlockNumber(blockNumber)
	require.Nil(t, err)
	assert.Equal(t, 5, len(txs.Txs))
}
