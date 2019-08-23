package sync

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"

	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/storage/state/statedb"
)

// HBTCSyncer syncs all HBTC external accounts
// HBTC account is the first ETH account of user
type HBTCSyncer struct {
	RPC               string
	symbol, ethSymbol string
	syncer            *ExternalSyncer
}

func newHBTCSyncer() *HBTCSyncer {
	h := &HBTCSyncer{symbol: "HBTC", ethSymbol: "ETH"}
	h.syncer = newExternalSyncer(h.symbol)

	return h
}

// GetExtBalance ...
func (hs *HBTCSyncer) GetExtBalance() error {
	// If ETH account exists
	ethAccount, ok := hs.syncer.Account.EBalances[hs.ethSymbol]
	if !ok {
		log.Warn().Msg("HBTC depends on ETH, but no ETH account available")
		return errors.New("ETH account does not exists")
	}

	hbtcAccount, ok := ethAccount[hs.syncer.Account.FirstExternalAddress[hs.ethSymbol]]
	if !ok {
		log.Warn().Msg("HBTC does not exist")
		return errors.New("HBTC account does not exists")
	}

	httpClient := newHTTPClient()
	resp, err := httpClient.Get(fmt.Sprintf("%s/%s", hs.RPC, hbtcAccount.Address))
	if err != nil {
		log.Error().Err(err).Msg("failed to get hbtc balance")
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read response body")
		return err
	}

	balance, err := strconv.ParseInt(strings.TrimSuffix(string(body), "\n"), 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse response body")
		return err
	}

	hs.syncer.ExtBalance[hbtcAccount.Address] = big.NewInt(balance)

	return nil
}

// Update updates accounts in cache as and when external balances
// external chains are updated.
func (hs *HBTCSyncer) Update() {
	if hs.syncer.Account.EBalances[hs.ethSymbol] == nil {
		log.Warn().Msg("No HBTC account available, skip")
		return
	}
	if hs.syncer.Account.EBalances[hs.symbol] == nil {
		hs.syncer.Account.EBalances[hs.symbol] = make(map[string]statedb.EBalance)
		hs.syncer.Account.EBalances[hs.symbol][hs.syncer.Account.FirstExternalAddress[hs.ethSymbol]] = statedb.EBalance{Address: hs.syncer.Account.FirstExternalAddress[hs.ethSymbol]}
	}

	// HBTC account is first ETH account of user.
	ethAccount := hs.syncer.Account.EBalances[hs.symbol][hs.syncer.Account.FirstExternalAddress[hs.ethSymbol]]
	hs.syncer.update(ethAccount.Address)
}
