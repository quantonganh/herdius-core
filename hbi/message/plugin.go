package message

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/herdius/herdius-core/accounts/account"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/storage/mempool"

	protoplugin "github.com/herdius/herdius-core/hbi/protobuf"
	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
)

type TxType int

const (
	Update TxType = iota
)

func (t TxType) String() string {
	return [...]string{"Update"}[t]
}

// BlockMessagePlugin will receive all Block specific messages.
type BlockMessagePlugin struct{ *network.Plugin }
type AccountMessagePlugin struct{ *network.Plugin }
type TransactionMessagePlugin struct{ *network.Plugin }

// Receive handles account specific messages
func (state *AccountMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {

	case *protoplugin.AccountRequest:
		getAccount(msg.Address, ctx)

	case *protoplugin.AccountResponse:
		log.Info().Msgf("Account Response: %v", msg)
	}
	return nil
}

// Receive handles block specific received messages
func (state *BlockMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {
	case *protoplugin.BlockHeightRequest:
		getBlock(msg.BlockHeight, ctx)

	case *protoplugin.BlockResponse:
		log.Info().Msgf("Block Response: %v", msg)
	}
	return nil
}

// Receive to handle transaction requests
func (state *TransactionMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {
	case *protoplugin.TxDetailRequest:

		txID := msg.GetTxId()
		log.Printf("Tx ID: %v", txID)
		getTx(txID, ctx)

	case *protoplugin.TxRequest:
		tx := msg.GetTx()
		accSrv := account.NewAccountService()
		account, err := accSrv.GetAccountByAddress(msg.Tx.GetSenderAddress())

		apiClient, err := ctx.Network().Client(ctx.Client().Address)

		if err != nil {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1,
				&protoplugin.TxResponse{
					TxId: "", Status: "failed", Queued: 0, Pending: 0,
					Message: "Couldn't find the account due to : " + err.Error(),
				})
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
			}
			return errors.New("Couldn't find the account due to: " + err.Error())
		}

		//Check Tx.Nonce > account.Nonce
		if account != nil {
			if !accSrv.VerifyAccountNonce(account, tx.GetAsset().Nonce) {
				failedVerificationMsg := "Transaction nonce (" + string(msg.Tx.GetAsset().Nonce) +
					") should be greater than account nonce (" + string(account.Nonce) + ")"
				err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1,
					&protoplugin.TxResponse{
						TxId: "", Status: "failed", Queued: 0, Pending: 0,
						Message: failedVerificationMsg,
					})
				if err != nil {
					return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
				}
				return errors.New(failedVerificationMsg)
			}
		}

		//Check if tx is of type account update

		update := Update.String()
		if strings.EqualFold(tx.Type, update) {
			postAccountUpdateTx(tx, ctx)
			return nil
		}

		// Check if asset has enough balance
		// account.Balance > Tx.Value
		if !accSrv.VerifyAccountBalance(account, tx.GetAsset().Value, tx.GetAsset().GetSymbol()) {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1,
				&protoplugin.TxResponse{
					TxId: "", Status: "failed", Queued: 0, Pending: 0,
					Message: "Not enough balance: " + string(msg.Tx.GetAsset().Value),
				})
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
			}
			return errors.New("Not enough balance: " + strconv.FormatUint(uint64(msg.Tx.GetAsset().Value), 10))
		}

		// Add Tx to Mempool
		mp := mempool.GetMemPool()
		txbz, err := cdc.MarshalJSON(tx)

		if err != nil {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1,
				&protoplugin.TxResponse{
					TxId: "", Status: "failed", Queued: 0, Pending: 0,
					Message: "Incorrect Transaction format : " + msg.Tx.GetSenderAddress(),
				})
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
			}
			return errors.New("Failed to Masshal Tx: " + msg.Tx.GetSenderAddress())
		}

		txCount := mp.AddTx(txbz)

		log.Info().Msgf("Remaining mempool txcount: %v", txCount)

		// Create the Transaction ID
		txID := cmn.CreateTxID(txbz)
		log.Info().Msgf("Tx ID : %v", txID)

		// Send Tx ID to client who sent the TX
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1,
			&protoplugin.TxResponse{
				TxId: txID, Status: "success", Queued: 0, Pending: 0,
			})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
		}
	}
	return nil
}

func postAccountUpdateTx(tx *protoplugin.Tx, ctx *network.PluginContext) error {
	// Add Tx to Mempool
	mp := mempool.GetMemPool()
	txbz, err := cdc.MarshalJSON(tx)
	apiClient, err := ctx.Network().Client(ctx.Client().Address)

	if err != nil {
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1,
			&protoplugin.TxResponse{
				TxId: "", Status: "failed", Queued: 0, Pending: 0,
				Message: "Transaction format incorrect : " + err.Error(),
			})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
		}
		return errors.New("Failed to Masshal Tx: " + err.Error())
	}

	txCount := mp.AddTx(txbz)

	log.Info().Msgf("Remaining mempool txcount: %v", txCount)

	// Create the Transaction ID
	txID := cmn.CreateTxID(txbz)
	log.Info().Msgf("Tx ID : %v", txID)

	// Send Tx ID to client who sent the TX
	err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1,
		&protoplugin.TxResponse{
			TxId: txID, Status: "success", Queued: 0, Pending: 0,
		})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
	}
	return nil
}

func getBlock(height int64, ctx *network.PluginContext) error {
	blockchainSvc := &blockchain.Service{}
	block, err := blockchainSvc.GetBlockByHeight(height)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to retreive the Block: :%v", err))
	}

	pc, err := ctx.Network().Client(ctx.Client().Address)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to get the peer client :%v", err))
	}
	supervisorAdd := ctx.Client().ID.Address

	if block.Header != nil {
		timestamp := &protoplugin.Timestamp{
			Nanos:   block.GetHeader().GetTime().GetNanos(),
			Seconds: block.GetHeader().GetTime().GetSeconds(),
		}

		totalTxs := block.GetHeader().TotalTxs

		blockRes := protoplugin.BlockResponse{
			BlockHeight:       block.GetHeader().GetHeight(),
			TotalTxs:          totalTxs,
			Time:              timestamp,
			SupervisorAddress: supervisorAdd,
		}

		log.Info().Msgf("Block Response at processor: %v", blockRes)

		err = pc.Reply(network.WithSignMessage(context.Background(), true), 1, &blockRes)

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
		}
		return nil
	}

	err = pc.Reply(network.WithSignMessage(context.Background(), true), 1, &protoplugin.BlockResponse{})

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
	}
	return nil
}
func getAccount(address string, ctx *network.PluginContext) error {
	accountSvc := &account.Service{}
	account, err := accountSvc.GetAccountByAddress(address)
	if err != nil {
		log.Error().Msgf("Failed to retreive the Account: %v", err)
	}

	apiClient, err := ctx.Network().Client(ctx.Client().Address)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to get the peer client :%v", err))
	}

	if account == nil {
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1, &protoplugin.AccountResponse{})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
		}
	}

	if account != nil {
		accountResp := protoplugin.AccountResponse{
			Address:     address,
			Nonce:       account.Nonce,
			Balance:     account.Balance,
			StorageRoot: account.StorageRoot,
			PublicKey:   account.PublicKey,
			Balances:    account.Balances,
		}

		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1, &accountResp)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
		}
	}
	return nil
}

func getTx(id string, ctx *network.PluginContext) error {
	txSvc := &blockchain.TxService{}
	txDetailRes, err := txSvc.GetTx(id)

	apiClient, err := ctx.Network().Client(ctx.Client().Address)

	if err != nil {
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1,
			&protoplugin.TxDetailResponse{})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
		}
		return errors.New("Failed due to: " + err.Error())
	}

	err = apiClient.Reply(network.WithSignMessage(context.Background(), true), 1, txDetailRes)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
	}
	return nil
}
