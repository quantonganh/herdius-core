package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/herdius/herdius-core/p2p/key"
	"github.com/herdius/herdius-core/supervisor/transaction"
)

func main() {
	GenAccounts()
	createTxs()
}

func createTxs() {

	file, _ := os.Create("txs-sec.json")
	defer file.Close()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parent := filepath.Dir(wd)
	for i := 1; i <= 10; i++ {
		filePath := filepath.Join(parent + "/testdata/secp205k1Accts/" + strconv.Itoa(i) + "_peer_id.json")

		nodeKey, _ := key.LoadOrGenNodeKey(filePath)

		for n := 1; n <= 300; n++ {
			msg := []byte("Transfer 10 BTC")

			pubKey := nodeKey.PubKey()
			sign, _ := nodeKey.PrivKey.Sign(msg)
			tx := transaction.Tx{
				Nonce:         uint64(n),
				Senderpubkey:  pubKey.Bytes(),
				Fee:           []byte("0"),
				Assetcategory: "Crypto",
				Assetname:     "BTC",
				Value:         []byte("10"),
				Signature:     sign,
				Message:       "Transfer 10 BTC",
			}

			buf := new(bytes.Buffer)
			encoder := json.NewEncoder(buf)
			encoder.Encode(tx)
			io.Copy(file, buf)
		}

	}

}

// GenAccounts ...
func GenAccounts() {
	fmt.Println("Creating secp256k1 curve specific accounts")
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parent := filepath.Dir(wd)
	for i := 1; i <= 10; i++ {
		filePath := filepath.Join(parent + "/testdata/secp205k1Accts/" + strconv.Itoa(i) + "_peer_id.json")

		nodeKey, err := key.LoadOrGenNodeKey(filePath)

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(nodeKey)
		}
	}

}
