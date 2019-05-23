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
	tokenAddress := common.HexToAddress(es.TokenContractAddress)

	instance, err := contract.NewToken(tokenAddress, client)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("es.RPC", es.RPC)
	address := common.HexToAddress(es.Account.Erc20Address)
	bal, err := instance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Token balance", bal)

	es.ExtBalance = bal
}

func (es *ERC20) Update() {
	es.Account.Balance = es.ExtBalance.Uint64()
	val := cache.AccountCache{Account: es.Account, LastExtBalance: es.ExtBalance}
	es.Cache.Set(es.Account.Address, val)
}
