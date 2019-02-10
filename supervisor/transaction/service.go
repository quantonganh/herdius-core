package transaction

// Service provides Transaction specific operations for the end users.
type Service interface {
	AddTx(Tx)
	GetTxList() *TxList
}
type service struct {
	txList TxList
}

// TxService creates a new transaction service
func TxService() Service {
	return &service{}
}

func (s *service) AddTx(tx Tx) {

	txList := s.GetTxList()
	transactions := (*txList).Transactions

	transactions = append(transactions, &tx)
	(*txList).Transactions = transactions
}

func (s *service) GetTxList() *TxList {
	return &s.txList
}
