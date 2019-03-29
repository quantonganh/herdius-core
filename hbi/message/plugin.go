package message

import (
	"context"
	"fmt"

	"github.com/herdius/herdius-core/accounts/account"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/storage/mempool"
	"github.com/herdius/herdius-core/tx"

	blockProtobuf "github.com/herdius/herdius-core/blockchain/protobuf"
	protoplugin "github.com/herdius/herdius-core/hbi/protobuf"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
)

// BlockMessagePlugin will receive all Block specific messages.
type BlockMessagePlugin struct{ *network.Plugin }
type AccountMessagePlugin struct{ *network.Plugin }
type TransactionMessagePlugin struct{ *network.Plugin }

var memPool = mempool.NewMemPool()

func (state *AccountMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {

	case *protoplugin.AccountRequest:
		accountSvc := &account.Service{}
		account, err := accountSvc.GetAccountByAddress(msg.GetAddress())
		if err != nil {
			log.Error().Msgf("Failed to retreive the Account: %v", err)
		}

		if account == nil {
			ctx.Network().BroadcastByAddresses(network.WithSignMessage(context.Background(), true), &blockProtobuf.ConnectionMessage{Message: "Account detail not found"}, "tcp://127.0.0.1:5555")
		}

		if account != nil {
			accountResp := protoplugin.AccountResponse{
				Address:     msg.GetAddress(),
				Nonce:       account.Nonce,
				Balance:     account.Balance,
				StorageRoot: account.StorageRoot,
			}
			ctx.Network().BroadcastByAddresses(network.WithSignMessage(context.Background(), true), &accountResp, "tcp://127.0.0.1:5555")

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
			Transactions:      500,
			Time:              timestamp,
			SupervisorAddress: supervisorAdd,
		}

		log.Info().Msgf("Block Response at processor: %v", blockRes)

		ctx.Network().BroadcastByAddresses(network.WithSignMessage(context.Background(), true), &blockRes, "tcp://127.0.0.1:5555")

	case *protoplugin.BlockResponse:
		log.Info().Msgf("Block Response: %v", msg)
	}
	return nil
}

func (state *TransactionMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {
	case *protoplugin.TransactionRequest:
		txsSrv := tx.GetTxsService()
		accSrv := account.NewAccountService()
		accAddress, err := accSrv.GetPublicAddress(msg.Tx.Senderpubkey)
		if err != nil || accAddress == "" {
			fmt.Println("couldn't find user account for specificied sender public key:", err)
			return err
		}
		senderAcc, err := accSrv.GetAccountByAddress(accAddress)
		if err != nil || senderAcc == nil {
			fmt.Println("couldn't get account by address; err:", err)
			return err
		}
		fmt.Println("senderAcc.Nonce, senderAcc.Balance:", senderAcc)

		err = txsSrv.ParseNewTxRequest(senderAcc.Nonce, senderAcc.Balance, msg)
		poolCount, err := memPool.AddTxs(txsSrv)
		if err != nil {
			log.Info().Msgf("error adding transaction to memory pool: ", err)
			return err
		}
		fmt.Println("poolCount in hbi:", poolCount)

	case *protoplugin.TransactionResponse:
		log.Info().Msgf("Transaction Response: %v", msg)
	}
	return nil
}
