package transaction

import (
	"bytes"
	"encoding/json"
)

const (
	empty = ""
	tab   = "\t"
)

// Tx : Transaction Detail
type Tx struct {
	Nonce         uint64 `json:"nonce"`
	Senderpubkey  []byte `json:"senderpubkey"`
	Fee           []byte `json:"fee"`
	Assetcategory string `json:"assetcategory"`
	Assetname     string `json:"assetname"`
	Value         []byte `json:"value"`
	Signature     []byte `json:"sign"`
	Message       string `json:"message"`
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
