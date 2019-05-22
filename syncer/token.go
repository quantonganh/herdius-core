package sync

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/herdius/herdius-core/storage/cache"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/herdius/herdius-core/syncer/contract"
)

type ERC20 struct {
	LastExtBalance       *big.Int
	ExtBalance           *big.Int
	Account              statedb.Account
	Cache                *cache.Cache
	TokenContractAddress string
	TokenSymbol          string
	RPC                  string
}

func (es *ERC20) GetExtBalance() {
	client, err := ethclient.Dial(es.RPC)
	if err != nil {
		log.Println("Error connecting ETH RPC", err)
	}
	tokenAddress := common.HexToAddress("0xB8c77482e45F1F44dE1745F52C74426C631bDD52")

	instance, err := contract.NewToken(tokenAddress, client)
	if err != nil {
		log.Fatal(err)
	}
	address := common.HexToAddress("0x4ef1dd38ace2ced41d4002d4aa7982d71c457001")
	bal, err := instance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Token balance", bal)

	es.ExtBalance = bal
}

func (es *ERC20) Update() {
	value, ok := es.Account.EBalances[es.TokenSymbol]
	if ok {
		value.UpdateBalance(es.ExtBalance.Uint64())
		es.Account.EBalances[es.TokenSymbol] = value
		val := cache.AccountCache{Account: es.Account, LastExtBalance: es.ExtBalance}
		es.Cache.Set(es.Account.Address, val)

	}

}
