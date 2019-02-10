package main

import (
	"net"

	"github.com/herdius/herdius-core/p2p/log"
)

func main() {
	// connect to this socket
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Error().Msgf("failed to connect %v:\n", err)
	} else {
		log.Info().Msgf("connection established %v\n", conn.RemoteAddr().String())

	}
}

/* package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	ed25519 "github.com/herdius/herdius-core/crypto/ed"
)

// Tx : ...
type Tx struct {
	Nonce         uint64 `json:"nonce"`
	Senderpubkey  []byte `json:"senderpubkey"`
	Fee           []byte `json:"fee"`
	Assetcategory string `json:"assetcategory"`
	Assetname     string `json:"assetname"`
	Value         []byte `json:"value"`
	Signature     []byte `json:"sign"`
}

func main() {
	fmt.Println("Starting the application...")

	for i := 1; i <= 10; i++ {
		msg := []byte("Transfer 10 BTC")
		privKey := ed25519.GenPrivKey()
		pubKey := privKey.PubKey()
		sign, _ := privKey.Sign(msg)
		tx := Tx{
			Nonce:         uint64(i),
			Senderpubkey:  pubKey.Bytes(),
			Fee:           []byte("100"),
			Assetcategory: "Crypto",
			Assetname:     "BTC",
			Value:         []byte("10"),
			Signature:     sign,
		}

		jsonValue, _ := json.Marshal(tx)

		response, err := http.Post("http://localhost:8080/tx", "application/json", bytes.NewBuffer(jsonValue))

		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			fmt.Printf("%v : Tx(%v)\n", string(data), i)
		}
	}

	response, err := http.Get("http://localhost:8080/txlist")

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Printf("Total number of TXs: %v\n", len(data))
	}

	fmt.Println("Terminating the user application...")
}
*/
