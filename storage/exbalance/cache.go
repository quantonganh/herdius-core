package exbalance

type BalanceStorage interface {
	Set(k string, x AccountCache)
	Get(k string) (AccountCache, bool)
	GetAll() map[string]AccountCache
	Close()
	CloseTest()
}
