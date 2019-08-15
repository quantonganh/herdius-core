package transaction

import (
	"bytes"
	"encoding/json"
)

const (
	empty = ""
	tab   = "\t"
)

// Asset ...
type Asset struct {
	Category                string `json:"category"`
	Symbol                  string `json:"symbol"`
	Network                 string `json:"network"`
	Value                   string `json:"value"`
	Fee                     string `json:"fee"`
	Nonce                   string `json:"nonce"`
	ExternalSenderAddress   string `json:"external_sender_address"`
	ExternalRecieverAddress string `json:"external_reciever_address"`
	ExternalNonce           uint64 `json:"external_nonce"`
	ExternalBlockHeight     uint64 `json:"external_block_height"`
	LockedAmount            uint64 `json:"locked_amount"`
	RedeemedAmount          uint64 `json:"redeemed_amount"`
}

// Tx ...
type Tx struct {
	SenderAddress   string `json:"sender_address"`
	SenderPubKey    string `json:"sender_pubkey"`
	ReceiverAddress string `json:"receiver_address"`
	Asset           Asset  `json:"asset"`
	Message         string `json:"message"`
	Signature       string `json:"sign"`
	Type            string `json:"type"`
	Status          string `json:"status"`
}

// TxList : List of Transactions
// Batches are to be created for 500-1000 Transactions
type TxList struct {
	Transactions []*Tx `json:"transactions"`
}

// PrettyPrint prints the json with readable indenting
func PrettyPrint(tx Tx) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(tx)
	if err != nil {
		return "Failed to parse to pretty print.", err
	}
	return buffer.String(), nil
}
