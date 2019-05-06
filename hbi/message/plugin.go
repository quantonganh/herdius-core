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

var nonce uint64 = 1

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
	case *protoplugin.TxsByAssetAndAddressRequest:
		address := msg.GetAddress()
		asset := msg.GetAsset()
		accSrv := account.NewAccountService()
		account, err := accSrv.GetAccountByAddress(address)
		apiClient, err := ctx.Network().Client(ctx.Client().Address)

		if err != nil {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxsResponse{})
			nonce++
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
			}
			return errors.New("Couldn't find the account due to: " + err.Error())
		}
		if account != nil && strings.EqualFold(account.Address, address) {
			getTxsByAssetAndAccount(asset, address, ctx)
		} else {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxsResponse{})
			nonce++
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
			}
			return nil
		}

	case *protoplugin.TxsByAddressRequest:
		address := msg.GetAddress()
		accSrv := account.NewAccountService()
		account, err := accSrv.GetAccountByAddress(address)
		apiClient, err := ctx.Network().Client(ctx.Client().Address)

		if err != nil {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxsResponse{})
			nonce++
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
			}
			return errors.New("Couldn't find the account due to: " + err.Error())
		}

		if account != nil &&
			strings.EqualFold(account.Address, address) {
			getTxs(address, ctx)
		} else {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxsResponse{})
			nonce++
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
			}
			return nil
		}

	case *protoplugin.TxDetailRequest:

		txID := msg.GetTxId()
		getTx(txID, ctx)

	case *protoplugin.TxRequest:
		tx := msg.GetTx()
		accSrv := account.NewAccountService()
		account, err := accSrv.GetAccountByAddress(msg.Tx.GetSenderAddress())

		apiClient, err := ctx.Network().Client(ctx.Client().Address)

		if err != nil {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxResponse{
					TxId: "", Status: "failed", Queued: 0, Pending: 0,
					Message: "Couldn't find the account due to : " + err.Error(),
				})
			nonce++
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
			}
			return errors.New("Couldn't find the account due to: " + err.Error())
		}

		//Check Tx.Nonce > account.Nonce
		if account != nil {
			if !accSrv.VerifyAccountNonce(account, tx.GetAsset().Nonce) {

				txNonce := strconv.FormatUint(msg.Tx.GetAsset().Nonce, 10)
				accountNonce := strconv.FormatUint(account.Nonce, 10)
				failedVerificationMsg := "Transaction nonce " + txNonce +
					" should be greater than account nonce " + accountNonce

				err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
					&protoplugin.TxResponse{
						TxId: "", Status: "failed", Queued: 0, Pending: 0,
						Message: failedVerificationMsg,
					})
				nonce++
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
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxResponse{
					TxId: "", Status: "failed", Queued: 0, Pending: 0,
					Message: "Not enough balance: " + string(msg.Tx.GetAsset().Value),
				})
			nonce++
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
			}
			return errors.New("Not enough balance: " + strconv.FormatUint(uint64(msg.Tx.GetAsset().Value), 10))
		}

		// Add Tx to Mempool
		mp := mempool.GetMemPool()
		txbz, err := cdc.MarshalJSON(tx)

		if err != nil {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxResponse{
					TxId: "", Status: "failed", Queued: 0, Pending: 0,
					Message: "Incorrect Transaction format : " + msg.Tx.GetSenderAddress(),
				})
			nonce++
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
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
			&protoplugin.TxResponse{
				TxId: txID, Status: "success", Queued: 0, Pending: 0,
			})
		nonce++
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
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
			&protoplugin.TxResponse{
				TxId: "", Status: "failed", Queued: 0, Pending: 0,
				Message: "Transaction format incorrect : " + err.Error(),
			})
		nonce++
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
	err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
		&protoplugin.TxResponse{
			TxId: txID, Status: "success", Queued: 0, Pending: 0,
		})
	nonce++
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

		err = pc.Reply(network.WithSignMessage(context.Background(), true), nonce, &blockRes)
		nonce++

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client :%v", err))
		}
		return nil
	}

	err = pc.Reply(network.WithSignMessage(context.Background(), true), nonce, &protoplugin.BlockResponse{})
	nonce++

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
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce, &protoplugin.AccountResponse{})
		nonce++
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

		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce, &accountResp)
		nonce++
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
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
			&protoplugin.TxDetailResponse{})
		nonce++
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
		}
		return errors.New("Failed due to: " + err.Error())
	}

	err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce, txDetailRes)
	nonce++
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
	}
	return nil
}

func getTxs(address string, ctx *network.PluginContext) error {
	txSvc := &blockchain.TxService{}
	txs, err := txSvc.GetTxs(address)

	apiClient, err := ctx.Network().Client(ctx.Client().Address)

	if err != nil {
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
			&protoplugin.TxDetailResponse{})
		nonce++
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
		}
		return errors.New("Failed due to: " + err.Error())
	}

	err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce, txs)
	nonce++
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
	}
	return nil
}

func getTxsByAssetAndAccount(asset, address string, ctx *network.PluginContext) error {
	txSvc := &blockchain.TxService{}
	txs, err := txSvc.GetTxsByAssetAndAddress(asset, address)

	apiClient, err := ctx.Network().Client(ctx.Client().Address)

	if err != nil {
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
			&protoplugin.TxDetailResponse{})
		nonce++
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
		}
		return errors.New("Failed due to: " + err.Error())
	}

	err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce, txs)
	nonce++
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
	}
	return nil
}
