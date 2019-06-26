package message

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/herdius/herdius-core/accounts/account"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/storage/mempool"

	"github.com/herdius/herdius-core/hbi/protobuf"
	protoplugin "github.com/herdius/herdius-core/hbi/protobuf"
	"github.com/herdius/herdius-core/libs/common"
	cmn "github.com/herdius/herdius-core/libs/common"
	plog "github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
)

type TxType int

const (
	Update TxType = iota
	Lock
)

var nonce uint64 = 1

func (t TxType) String() string {
	return [...]string{"Update", "Lock"}[t]
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
		plog.Info().Msgf("Account Response: %v", msg)
	}
	return nil
}

// Receive handles block specific received messages
func (state *BlockMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {
	case *protoplugin.BlockHeightRequest:
		getBlock(msg.BlockHeight, ctx)

	case *protoplugin.BlockResponse:
		plog.Info().Msgf("Block Response: %v", msg)
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
		accSrv.SetReceiverAddress(tx.RecieverAddress)
		accSrv.SetAssetSymbol(tx.Asset.Symbol)
		accSrv.SetExtAddress(tx.Asset.ExternalSenderAddress)
		accSrv.SetTxValue(tx.Asset.Value)
		accSrv.SetTxLockedAmount(tx.Asset.LockedAmount)
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
		accSrv.SetAccount(account)
		// Check if tx is of type account update
		// and verify external address exists
		update := Update.String()
		if strings.EqualFold(tx.Type, update) {
			if accSrv.AccountExternalAddressExist() {
				failedVerificationMsg := "External account existed: " + tx.Asset.ExternalSenderAddress
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
			if accSrv.AccountEBalancePerAssetReachLimit() {
				failedVerificationMsg := "Account reached number of addresses limit"
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
			postAccountUpdateTx(tx, ctx, accSrv)
			return nil
		}
		// Check if tx is of type lock
		// verify if external account address doesn't exists
		// verify if receiver address is herdius zero address
		lock := Lock.String()
		if strings.EqualFold(tx.Type, lock) {
			if !accSrv.AccountExternalAddressExist() {
				failedVerificationMsg := "External address does not exist: " + tx.Asset.ExternalSenderAddress
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
			if !accSrv.IsHerdiusZeroAddress() {
				failedVerificationMsg := "Incorrect herdius zero address: " + tx.RecieverAddress
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
			if !accSrv.VerifyLockedAmount() {
				failedVerificationMsg := "Account does not have enough locked amount"
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
		// Check if asset has enough balance
		// account.Balance > Tx.Value
		if !strings.EqualFold(tx.Type, lock) && !accSrv.VerifyAccountBalance() {
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
		log.Println("Add tx to mempool")
		pending, queue := mp.AddTx(tx, accSrv)

		plog.Info().Msgf("Remaining mempool pending, queue: %+v %+v", pending, queue)

		// Create the Transaction ID
		txID := cmn.CreateTxID(txbz)
		plog.Info().Msgf("Tx ID : %v", txID)

		// Send Tx ID to client who sent the TX
		err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
			&protoplugin.TxResponse{
				TxId: txID, Status: "success", Queued: int64(queue), Pending: int64(pending),
			})
		nonce++
		if err != nil {
			return fmt.Errorf("Failed to reply to client :%v", err)
		}
	case *protoplugin.TxUpdateRequest:
		log.Println("Update request received")
		apiClient, err := ctx.Network().Client(ctx.Client().Address)
		if err != nil {
			return fmt.Errorf("can't find requesting API client: %v", apiClient)
		}
		log.Println("Request ingress from API node IP:", apiClient.Address)
		id := msg.GetTxId()
		newTx := msg.GetTx()
		log.Println("Processing request to update Tx, ID:", id)
		updatedID, updatedTx, err := putTxUpdateRequest(id, newTx)
		if err != nil {
			errRep := apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxUpdateResponse{
					Error:  err.Error(),
					Status: false,
				})
			if errRep != nil {
				return fmt.Errorf("could not reply to API client, transaction not updated: %v", errRep)
			}
			nonce++
			return fmt.Errorf("could not update request: %v", err)
		}
		errRep := apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
			&protoplugin.TxUpdateResponse{
				Status: true,
				TxId:   updatedID,
				Tx:     updatedTx,
			})
		if errRep != nil {
			return fmt.Errorf("could not reply to API client, but transaction was updated: %v", errRep)
		}
		nonce++
		return nil
	case *protoplugin.TxDeleteRequest:
		mp := mempool.GetMemPool()
		succ := mp.DeleteTx(msg.TxId)
		apiClient, err := ctx.Network().Client(ctx.Client().Address)
		if err != nil {
			return fmt.Errorf("Failed to get API client: %v", err)
		}
		if succ {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxUpdateResponse{
					TxId:   msg.TxId,
					Status: succ,
				})
		} else {
			err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce,
				&protoplugin.TxUpdateResponse{
					Status: succ,
					Error:  fmt.Sprintf("Unable to find Tx (id: %v) in memory pool", msg.TxId),
				})
		}
		nonce++
		if err != nil {
			return fmt.Errorf("Failed to reply to client :%v", err)
		}
	case *protoplugin.TxLockedRequest:
		log.Println("`locked` GET request received")
		getLockedTxs(ctx)
	}
	return nil
}

func postAccountUpdateTx(tx *protoplugin.Tx, ctx *network.PluginContext, as account.ServiceI) error {
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

	pending, queue := mp.AddTx(tx, as)

	plog.Info().Msgf("Remaining mempool pending, queue: %+v %+v", pending, queue)

	// Create the Transaction ID
	txID := cmn.CreateTxID(txbz)
	plog.Info().Msgf("Tx ID : %v", txID)

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

		plog.Info().Msgf("Block Response at processor: %v", blockRes)

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
		plog.Error().Msgf("Failed to retreive the Account: %v", err)
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
		eBalances := make(map[string]*protoplugin.EBalanceAsset)

		for asset, assetAccount := range account.EBalances {
			eBalances[asset] = &protoplugin.EBalanceAsset{}
			eBalances[asset].Asset = make(map[string]*protobuf.EBalance)
			for _, eb := range assetAccount.Asset {
				eBalanceRes := &protoplugin.EBalance{
					Address:         eb.Address,
					Balance:         eb.Balance,
					LastBlockHeight: eb.LastBlockHeight,
					Nonce:           eb.Nonce,
				}
				eBalances[asset].Asset[eb.Address] = eBalanceRes
			}
		}
		accountResp := protoplugin.AccountResponse{
			Address:         address,
			Nonce:           account.Nonce,
			Balance:         account.Balance,
			StorageRoot:     account.StorageRoot,
			PublicKey:       account.PublicKey,
			EBalances:       eBalances,
			Erc20Address:    account.Erc20Address,
			ExternalNonce:   account.ExternalNonce,
			LastBlockHeight: account.LastBlockHeight,
		}
		fmt.Println("Asked for account detail, ", accountResp)
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

// putTxUpdateRequest upates the Tx with the input string. After updating, calculates new Tx ID
func putTxUpdateRequest(id string, newTx *protoplugin.Tx) (string, *protoplugin.Tx, error) {
	mp := mempool.GetMemPool()
	i, origTx, err := mp.GetTx(id)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get original transaction details (id: %v), err: %v", id, err)
	}
	if origTx == nil {
		return "", nil, fmt.Errorf("requested Tx (id: %v) does not exist in memory pool; it may have been flushed from the memory pool into a block", id)
	}
	updatedTx, err := mp.UpdateTx(i, newTx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to update Tx in MemPool with new values: %v", err)
	}
	updatedBz, err := cdc.MarshalJSON(updatedTx)
	if err != nil {
		return "", nil, fmt.Errorf("could not marshal updated transaction back into memory pool: %v", err)
	}
	newID := common.CreateTxID(updatedBz)
	return newID, updatedTx, nil
}

func getLockedTxs(ctx *network.PluginContext) error {
	txSvc := &blockchain.TxService{}
	txs, err := txSvc.GetLockedTxs()
	if err != nil {
		return fmt.Errorf("error getting lost transactions: %v", err)
	}

	apiClient, err := ctx.Network().Client(ctx.Client().Address)
	if err != nil {
		return fmt.Errorf("error getting API client: %v", err)
	}
	err = apiClient.Reply(network.WithSignMessage(context.Background(), true), nonce, txs)
	nonce++
	if err != nil {
		return fmt.Errorf("error replying to apiClient: %v", err)
	}
	return nil
}
