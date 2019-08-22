package sync

import (
	"math/big"
	"os"
	"testing"

	"github.com/herdius/herdius-core/storage/db"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/stretchr/testify/assert"
)

func TestNoResponseFromAPI(t *testing.T) {
	var accountCache external.BalanceStorage
	badgerdb := db.NewDB("test.syncdb", db.GoBadgerBackend, "test.syncdb")
	accountCache = external.NewDB(badgerdb)
	defer func() {
		badgerdb.Close()
		os.RemoveAll("./test.syncdb")
	}()

	addr := "BTC-1"
	eBalances := make(map[string]map[string]statedb.EBalance)
	eBalances["BTC"] = make(map[string]statedb.EBalance)
	eBalances["BTC"][addr] = statedb.EBalance{Address: addr, Balance: uint64(8)}

	account := statedb.Account{}
	account.EBalances = eBalances
	account.Address = "testBTCAddress"

	bs := newBTCSyncer()
	bs.syncer.Account = account
	bs.syncer.Storage = accountCache
	bs.syncer.ExtBalance[addr] = big.NewInt(1)
	bs.syncer.Nonce[addr] = 0
	bs.syncer.BlockHeight[addr] = big.NewInt(0)
	bs.Update()

	bs.syncer.ExtBalance[addr] = nil
	assert.Panics(t, bs.Update, "")

}
