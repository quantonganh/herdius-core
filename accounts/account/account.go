package account

// Account : Account Detail
type Account struct {
	Nonce       uint64 `json:"nonce"`
	Address     string `json:"address"`
	Balance     []byte `json:"balance"`
	StorageRoot []byte `json:"storageRoot"`
}
