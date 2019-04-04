package message

import (
	"context"
	"errors"
	"strings"

	"github.com/herdius/herdius-core/accounts/account"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/storage/mempool"

	blockProtobuf "github.com/herdius/herdius-core/blockchain/protobuf"
	protoplugin "github.com/herdius/herdius-core/hbi/protobuf"
	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
)

var (
	clientAddress = "tcp://127.0.0.1:5555"
)

// BlockMessagePlugin will receive all Block specific messages.
type BlockMessagePlugin struct{ *network.Plugin }
type AccountMessagePlugin struct{ *network.Plugin }
type TransactionMessagePlugin struct{ *network.Plugin }

func (state *AccountMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {

	case *protoplugin.AccountRequest:
		accountSvc := &account.Service{}
		account, err := accountSvc.GetAccountByAddress(msg.GetAddress())
		if err != nil {
			log.Error().Msgf("Failed to retreive the Account: %v", err)
		}

		if account == nil {
			ctx.Network().BroadcastByAddresses(network.WithSignMessage(context.Background(), true), &blockProtobuf.ConnectionMessage{Message: "Account detail not found"}, clientAddress)
		}

		if account != nil {
			accountResp := protoplugin.AccountResponse{
				Address:     msg.GetAddress(),
				Nonce:       account.Nonce,
				Balance:     account.Balance,
				StorageRoot: account.StorageRoot,
			}
			ctx.Network().BroadcastByAddresses(network.WithSignMessage(context.Background(), true), &accountResp, clientAddress)

		}

	case *protoplugin.AccountResponse:
		log.Info().Msgf("Account Response: %v", msg)
	}
	return nil
}

// Receive handles block specific received messages
func (state *BlockMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {
	case *protoplugin.BlockHeightRequest:
		blockchainSvc := &blockchain.Service{}
		block, err := blockchainSvc.GetBlockByHeight(msg.GetBlockHeight())

		if err != nil {
			log.Error().Msgf("Failed to retreive the Block: %v", err)
		}

		log.Info().Msgf("Block height at the blockMessage plugin%v", msg.GetBlockHeight())

		supervisorAdd := ctx.Client().ID.Address

		log.Info().Msgf("Supervisor Address: %s", supervisorAdd)

		timestamp := &protoplugin.Timestamp{
			Nanos:   block.GetHeader().GetTime().GetNanos(),
			Seconds: block.GetHeader().GetTime().GetSeconds(),
		}

		blockRes := protoplugin.BlockResponse{
			BlockHeight:       block.GetHeader().GetHeight(),
			TotalTxs:          block.GetHeader().TotalTxs,
			Time:              timestamp,
			SupervisorAddress: supervisorAdd,
		}

		log.Info().Msgf("Block Response at processor: %v", blockRes)

		ctx.Network().BroadcastByAddresses(network.WithSignMessage(context.Background(), true), &blockRes, clientAddress)

	case *protoplugin.BlockResponse:
		log.Info().Msgf("Block Response: %v", msg)
	}
	return nil
}

// Receive to handle transaction requests
func (state *TransactionMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {
	case *protoplugin.TxRequest:
		tx := msg.GetTx()

		accSrv := account.NewAccountService()

		account, err := accSrv.GetAccountByAddress(msg.Tx.GetSenderAddress())
		if err != nil {
			log.Error().Msgf("Couldn't find the account for: %v", msg.Tx.GetSenderAddress())
		}

		//Check Tx.Nonce > account.Nonce
		if !accSrv.VerifyAccountNonce(account, tx.GetAsset().Nonce) {
			return errors.New("Incorrect Transaction Nonce: " + string(msg.Tx.GetAsset().Nonce))
		}

		// Check if asset is HER Token, then check
		// account.Balance > Tx.Value
		if strings.EqualFold(tx.GetAsset().Symbol, "HER") && !accSrv.VerifyAccountBalance(account, tx.GetAsset().Value) {
			return errors.New("Not enough HER Tokens: " + msg.Tx.GetSenderAddress())
		}

		// Add Tx to Mempool
		mp := mempool.GetMemPool()
		txbz, err := cdc.MarshalJSON(tx)

		if err != nil {
			return errors.New("Failed to Masshal Tx: " + msg.Tx.GetSenderAddress())
		}

		txCount := mp.AddTx(txbz)

		log.Info().Msgf("Remaining mempool txcount: %v", txCount)

		// Create the Transaction ID
		txID := cmn.CreateTxID(txbz)
		log.Info().Msgf("Tx ID : %v", txID)

		// Send Tx ID to client who sent the TX
		ctx.Network().BroadcastByAddresses(network.WithSignMessage(context.Background(), true),
			&protoplugin.TxResponse{
				TxId: txID, Status: "success", Queued: 0, Pending: 0,
			}, clientAddress)

	}
	return nil
}
