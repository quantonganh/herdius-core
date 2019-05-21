package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	b64 "encoding/base64"
	"os"
	"sync"

	pluginproto "github.com/herdius/herdius-core/hbi/protobuf"

	"github.com/ethereum/go-ethereum/common"
	"github.com/herdius/herdius-core/blockchain/protobuf"
	hehash "github.com/herdius/herdius-core/crypto/herhash"
	"github.com/herdius/herdius-core/crypto/merkle"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	cmn "github.com/herdius/herdius-core/libs/common"
	cryptokeys "github.com/herdius/herdius-core/p2p/crypto"
	"github.com/herdius/herdius-core/p2p/key"
	plog "github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"

	"github.com/herdius/herdius-core/storage/cache"
	"github.com/herdius/herdius-core/storage/mempool"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/herdius/herdius-core/supervisor/transaction"
	txbyte "github.com/herdius/herdius-core/tx"
)

// SupervisorI is an interface
type SupervisorI interface {
	AddValidator(publicKey []byte, address string) error
	RemoveValidator(address string)
	CreateChildBlock(net *network.Network, txs *transaction.TxList, height int64, previousBlockHash []byte) *protobuf.ChildBlock
	SetWriteMutex()
	GetChildBlockMerkleHash() ([]byte, error)
	GetValidatorGroupHash() ([]byte, error)
	GetNextValidatorGroupHash() ([]byte, error)
	CreateBaseBlock(lastBlock *protobuf.BaseBlock) (*protobuf.BaseBlock, error)
	GetMutex() *sync.Mutex
	ProcessTxs(lastBlock *protobuf.BaseBlock, net *network.Network, waitTime, noOfPeersInGroup int, stateRoot []byte) (*protobuf.BaseBlock, error)
	ShardToValidators(*txbyte.Txs, *network.Network, []byte) error
}

var (
	_ SupervisorI = (*Supervisor)(nil)
)

// Supervisor is concrete implementation of SupervisorI
type Supervisor struct {
	TxBatches           *[]txbyte.Txs // TxGroups will consist of list of the transaction batches
	writerMutex         *sync.Mutex
	ChildBlock          []*protobuf.ChildBlock
	Validator           []*protobuf.Validator
	ValidatorChildblock map[string]*protobuf.BlockID //Validator address pointing to child block hash
	VoteInfoData        map[string][]*protobuf.VoteInfo
	StateRoot           []byte
	memPoolChan         chan<- mempool.MemPool
}

//GetMutex ...
func (s *Supervisor) GetMutex() *sync.Mutex {
	return s.writerMutex
}

// AddValidator adds a validator to group
func (s *Supervisor) AddValidator(publicKey []byte, address string) error {
	if s.Validator == nil {
		s.Validator = make([]*protobuf.Validator, 0)
	}
	validator := &protobuf.Validator{
		Address:      address,
		PubKey:       publicKey,
		Stakingpower: 100,
	}
	//s.writerMutex.Lock()
	s.Validator = append(s.Validator, validator)
	//s.writerMutex.Unlock()
	return nil
}

// RemoveValidator removes a validator from validators list
func (s *Supervisor) RemoveValidator(address string) {
	for i, v := range s.Validator {

		if v.Address == address {
			s.writerMutex.Lock()
			s.Validator[i] = s.Validator[len(s.Validator)-1]
			// We do not need to put s.Validator[i] at the end, as it will be discarded anyway
			s.Validator = s.Validator[:len(s.Validator)-1]
			s.writerMutex.Unlock()
			break
		}
	}

}

// SetWriteMutex ...
func (s *Supervisor) SetWriteMutex() {
	s.writerMutex = new(sync.Mutex)
}

// GetChildBlockMerkleHash creates merkle hash of all the child blocks
func (s *Supervisor) GetChildBlockMerkleHash() ([]byte, error) {
	//cdc.MarshalBinaryBare()
	if s.ChildBlock != nil && len(s.ChildBlock) > 0 {
		cbBzs := make([][]byte, len(s.ChildBlock))

		for i := 0; i < len(s.ChildBlock); i++ {
			cb := s.ChildBlock[i]
			cbBz, err := cdc.MarshalBinaryBare(*cb)
			if err != nil {
				return nil, fmt.Errorf(fmt.Sprintf("Child block Marshaling failed: %v.", err))
			}
			cbBzs[i] = cbBz
		}

		return merkle.SimpleHashFromByteSlices(cbBzs), nil
	}
	return nil, fmt.Errorf("no Child block available: %v", s.ChildBlock)
}

// GetValidatorGroupHash creates merkle hash of all the validators
func (s *Supervisor) GetValidatorGroupHash() ([]byte, error) {
	if s.Validator != nil && len(s.Validator) > 0 {
		vlBzs := make([][]byte, len(s.Validator))

		for i := 0; i < len(s.Validator); i++ {
			vl := s.Validator[i]
			vlBz, err := cdc.MarshalBinaryBare(*vl)
			if err != nil {
				return nil, fmt.Errorf(fmt.Sprintf("Validator Marshaling failed: %v.", err))
			}
			vlBzs[i] = vlBz
		}

		return merkle.SimpleHashFromByteSlices(vlBzs), nil
	}
	return nil, fmt.Errorf(fmt.Sprintf("No Child block available: %v.", s.ChildBlock))
}

// GetNextValidatorGroupHash ([]byte, error) creates merkle hash of all the next validators
func (s *Supervisor) GetNextValidatorGroupHash() ([]byte, error) {
	if s.Validator != nil && len(s.Validator) > 0 {
		vlBzs := make([][]byte, len(s.Validator))

		for i := 0; i < len(s.Validator); i++ {
			vl := s.Validator[i]
			vlBz, err := cdc.MarshalBinaryBare(*vl)
			if err != nil {
				return nil, fmt.Errorf(fmt.Sprintf("Validator Marshaling failed: %v.", err))
			}
			vlBzs[i] = vlBz
		}

		return merkle.SimpleHashFromByteSlices(vlBzs), nil
	}
	return nil, fmt.Errorf(fmt.Sprintf("No Child block available: %v.", s.ChildBlock))
}

// CreateBaseBlock creates the base block with all the child blocks
func (s *Supervisor) CreateBaseBlock(lastBlock *protobuf.BaseBlock) (*protobuf.BaseBlock, error) {
	// Create the merkle hash of all the child blocks
	cbMerkleHash, err := s.GetChildBlockMerkleHash()
	if err != nil {
		plog.Error().Msgf("Failed to create Merkle Hash of Validators: %v", err)
	}

	// Create the merkle hash of all the validators
	vgHash, err := s.GetValidatorGroupHash()
	if err != nil {
		plog.Error().Msgf("Failed to create Merkle Hash of Validators: %v", err)
	}

	// Create the merkle hash of all the next validators
	nvgHash, err := s.GetNextValidatorGroupHash()
	if err != nil {
		plog.Error().Msgf("Failed to create Merkle Hash of Next Validators: %v", err)
	}

	height := lastBlock.GetHeader().GetHeight()

	// create array of vote commits
	votecommits := make([]protobuf.VoteCommit, 0)
	for _, v := range s.ChildBlock {
		var cbh cmn.HexBytes
		cbh = v.GetHeader().GetBlockID().GetBlockHash()
		groupVoteInfo := s.VoteInfoData[cbh.String()]

		voteCommit := protobuf.VoteCommit{
			BlockID: v.GetHeader().GetBlockID(),
			Vote:    groupVoteInfo,
		}
		votecommits = append(votecommits, voteCommit)
	}

	vcbz, err := cdc.MarshalJSON(votecommits)
	if err != nil {
		plog.Error().Msgf("Vote commits marshaling failed.: %v", err)
	}

	ts := time.Now().UTC()
	baseHeader := &protobuf.BaseHeader{
		Block_ID:               &protobuf.BlockID{},
		LastBlockID:            lastBlock.GetHeader().GetBlock_ID(),
		Height:                 height + 1,
		ValidatorGroupHash:     vgHash,
		NextValidatorGroupHash: nvgHash,
		ChildBlockHash:         cbMerkleHash,
		LastVoteHash:           vcbz,
		StateRoot:              s.StateRoot,
		Time: &protobuf.Timestamp{
			Seconds: ts.Unix(),
			Nanos:   ts.UnixNano(),
		},
	}

	blockHashBz, err := cdc.MarshalJSON(baseHeader)
	if err != nil {
		plog.Error().Msgf("Base Header marshaling failed.: %v", err)
	}
	blockHash := hehash.Sum(blockHashBz)

	baseHeader.GetBlock_ID().BlockHash = blockHash

	childBlocksBz, err := cdc.MarshalJSON(s.ChildBlock)
	if err != nil {
		plog.Error().Msgf("Child blocks marshaling failed.: %v", err)
	}

	// Vote commits marshalling
	vcBz, err := cdc.MarshalJSON(votecommits)
	if err != nil {
		plog.Error().Msgf("Vote Commits marshaling failed.: %v", err)
	}

	// Validators marshaling
	valsBz, err := cdc.MarshalJSON(s.Validator)
	if err != nil {
		plog.Error().Msgf("Validators marshaling failed.: %v", err)
	}
	s.writerMutex.Lock()
	baseBlock := &protobuf.BaseBlock{
		Header:        baseHeader,
		ChildBlock:    childBlocksBz,
		VoteCommits:   vcBz,
		Validator:     valsBz,
		NextValidator: valsBz,
	}
	s.writerMutex.Unlock()
	return baseBlock, nil
}

// CreateChildBlock creates an initial child block
func (s *Supervisor) CreateChildBlock(net *network.Network, txs *transaction.TxList, height int64, previousBlockHash []byte) *protobuf.ChildBlock {
	txList := *txs
	if len(txList.Transactions) == 0 {
		return nil
	}

	numTxs := len(txList.Transactions)

	txbzs := make([][]byte, 0)
	txservice := txbyte.GetTxsService()

	for _, tx := range txList.Transactions {

		txbz, err := cdc.MarshalJSON(*tx)
		//plog.Fatalf("Marshalling failed: %v.", err)
		if err != nil {
			return nil
		}
		txbzs = append(txbzs, txbz)
	}
	txservice.SetTxs(txbzs)

	// Get Merkle Root Hash of all transactions
	rootHash := txservice.MerkleHash()

	//Supervisor details
	var keys *cryptokeys.KeyPair
	pubKey := make([]byte, 0)
	var address string
	if net != nil {
		keys = net.GetKeys()
		pubKey = keys.PubKey.Bytes()
		address = keys.PubKey.GetAddress()
	}

	// TODO: Id value calculation needs to implemented.
	id := &protobuf.ID{
		PublicKey: pubKey,
		Address:   address,
		Id:        []byte{0},
	}

	lastBlockID := &protobuf.BlockID{
		BlockHash: previousBlockHash,
	}
	// Create the child block
	header := &protobuf.Header{
		SupervisorID: id,
		NumTxs:       int64(numTxs),
		TotalTxs:     int64(numTxs),
		RootHash:     rootHash,
		Height:       height,
		LastBlockID:  lastBlockID,
	}
	hbz, _ := cdc.MarshalJSON(header)

	// Create the SHA256 value of the header
	// SHA256 Block Hash value is calculated using below header details:
	// Supervisor ID, # of txs, total txs and root hash
	// TODO: Need to make it better
	blockhash := hehash.Sum(hbz)

	blockID := &protobuf.BlockID{
		BlockHash: blockhash,
	}

	header.BlockID = blockID
	txsData := &protobuf.TxsData{
		Tx: txbzs,
	}
	s.writerMutex.Lock()
	cb := &protobuf.ChildBlock{
		Header:  header,
		TxsData: txsData,
	}
	s.writerMutex.Unlock()
	return cb
}

// ProcessTxs will process transactions.
// It will check whether to send the transactions to Validators
// or to be included in Singular base block
func (s *Supervisor) ProcessTxs(lastBlock *protobuf.BaseBlock, net *network.Network, waitTime, noOfPeersInGroup int, stateRoot []byte) (*protobuf.BaseBlock, error) {
	mp := mempool.GetMemPool()
	txs := mp.GetTxs()

	select {
	//case <-time.After(waitTime * time.Second):
	case <-time.After(time.Duration(waitTime) * time.Second):
		if len(*txs) <= 0 {
			log.Printf("Block creation wait time (%ds) elapsed, but no transaction to process", waitTime)
			return nil, nil
		}
		log.Printf("Block creation wait time (%d) elapsed, creating block", waitTime)
		// TODO once Validator Group capabilities developed, only when there are 2+ Validators should we Shard to Validators.
		if len(s.Validator) <= 0 {
			baseBlock, err := s.createSingularBlock(lastBlock, net, *txs, mp, stateRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to create base block: %v", err)
			}
			mp.RemoveTxs(len(*txs))
			return baseBlock, nil
		}
		err := s.ShardToValidators(txs, net, stateRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to shard Txs to child blocks: %v", err)
		}
		mp.RemoveTxs(len(*txs))
	}
	return nil, nil
}

func (s *Supervisor) createSingularBlock(lastBlock *protobuf.BaseBlock, net *network.Network, txs txbyte.Txs, mp *mempool.MemPool, stateRoot []byte) (*protobuf.BaseBlock, error) {
	stateTrie, err := statedb.NewTrie(common.BytesToHash(stateRoot))
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to retrieve the state trie: %v.", err))
	}

	updateAccs := accountCache.GetAll()

	log.Println("Total Accounts to update", len(updateAccs))
	for address, account := range updateAccs {

		switch v := account.Object.(type) {

		case cache.AccountCache:
			{

				account := account.Object.(cache.AccountCache).Account
				log.Println("Updating Account from cache", account)

				sactbz, err := cdc.MarshalJSON(account)
				if err != nil {
					plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
					continue
				}
				stateTrie.TryUpdate([]byte(address), sactbz)
			}
		default:
			{
				log.Println("Failed to get cache", v)

			}
		}
	}

	for i, txbz := range txs {

		var tx pluginproto.Tx
		err := cdc.UnmarshalJSON(txbz, &tx)
		if err != nil {
			log.Printf("Failed to Unmarshal tx: %v", err)
			continue
		}

		// Get the public key of the sender
		senderAddress := tx.GetSenderAddress()
		pubKeyS, err := b64.StdEncoding.DecodeString(tx.GetSenderPubkey())
		if err != nil {
			plog.Error().Msgf("Failed to decode sender public key: %v", err)
			tx.Status = "failed"
			txbz, err = cdc.MarshalJSON(&tx)
			txs[i] = txbz
			if err != nil {
				plog.Error().Msgf("Failed to encode failed tx: %v", err)
			}
			continue
		}

		var pubKey secp256k1.PubKeySecp256k1
		copy(pubKey[:], pubKeyS)

		// Verify the signature
		// if verification failed update the tx status as failed tx.
		//Recreate the TX
		asset := &pluginproto.Asset{
			Category:              tx.Asset.Category,
			Symbol:                tx.Asset.Symbol,
			Network:               tx.Asset.Network,
			Value:                 tx.Asset.Value,
			Fee:                   tx.Asset.Fee,
			Nonce:                 tx.Asset.Nonce,
			ExternalSenderAddress: tx.Asset.ExternalSenderAddress,
		}
		verifiableTx := pluginproto.Tx{
			SenderAddress:   tx.SenderAddress,
			SenderPubkey:    tx.SenderPubkey,
			RecieverAddress: tx.RecieverAddress,
			Asset:           asset,
			Message:         tx.Message,
			Type:            tx.Type,
		}

		txbBeforeSign, err := json.Marshal(verifiableTx)

		if err != nil {
			plog.Error().Msgf("Failed to marshal the transaction to verify sign: %v", err)
			continue
		}

		decodedSig, err := b64.StdEncoding.DecodeString(tx.Sign)

		if err != nil {
			plog.Error().Msgf("Failed to decode the base64 sign to verify sign: %v", err)
			continue
		}

		signVerificationRes := pubKey.VerifyBytes(txbBeforeSign, decodedSig)
		if !signVerificationRes {
			plog.Error().Msgf("Signature Verification Failed: %v", signVerificationRes)
			tx.Status = "failed"
			txbz, err = cdc.MarshalJSON(&tx)
			txs[i] = txbz
			if err != nil {
				plog.Error().Msgf("Failed to encode failed tx: %v", err)
			}
			continue
		}
		var senderAccount statedb.Account
		senderAddressBytes := []byte(senderAddress)

		// Get account details from state trie
		senderActbz, err := stateTrie.TryGet(senderAddressBytes)
		if err != nil {
			plog.Error().Msgf("Failed to retrieve account detail: %v", err)
			continue
		}

		if len(senderActbz) > 0 {
			err = cdc.UnmarshalJSON(senderActbz, &senderAccount)
			if err != nil {
				plog.Error().Msgf("Failed to Unmarshal account: %v", err)
				continue
			}
		}

		// Check if tx is of type account update
		if strings.EqualFold(tx.Type, "External") {
			symbol := tx.Asset.Symbol
			if symbol != "BTC" && symbol != "ETH" {
				plog.Error().Msgf("Unsupported external asset symbol: %v", symbol)
				continue
			}

			// By default each of the new accounts will have HER token (with balance 0)
			// added to the map object balances
			balance := senderAccount.EBalances[symbol]
			if balance == (statedb.EBalance{}) {
				plog.Error().Msgf("Sender has no assets for the given symbol: %v", symbol)
				continue
			}
			if balance.Balance <= tx.Asset.Value {
				plog.Error().Msgf("Sender does not have enough assets in account (%d) to send transaction amount (%d)", balance.Balance, tx.Asset.Value)
				continue
			}
			balance.Balance -= tx.Asset.Value

			senderAccount.Nonce = tx.Asset.Nonce
			senderAccount.EBalances[symbol] = balance

			sactbz, err := cdc.MarshalJSON(senderAccount)
			if err != nil {
				plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
				continue
			}
			addressBytes := []byte(pubKey.GetAddress())
			err = stateTrie.TryUpdate(addressBytes, sactbz)
			if err != nil {
				plog.Error().Msgf("Failed to store account in state db: %v", err)
				tx.Status = "failed"
				txbz, err = cdc.MarshalJSON(&tx)
				txs[i] = txbz
				if err != nil {
					plog.Error().Msgf("Failed to encode failed tx: %v", err)
				}
			}
			tx.Status = "success"
			txbz, err = cdc.MarshalJSON(&tx)
			txs[i] = txbz
			if err != nil {
				plog.Error().Msgf("Failed to encode failed tx: %v", err)
			}

			continue

		} else if strings.EqualFold(tx.Type, "Update") {
			senderAccount = *(updateAccount(&senderAccount, &tx))

			sactbz, err := cdc.MarshalJSON(senderAccount)
			if err != nil {
				plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
				continue
			}
			addressBytes := []byte(pubKey.GetAddress())
			err = stateTrie.TryUpdate(addressBytes, sactbz)
			if err != nil {
				plog.Error().Msgf("Failed to store account in state db: %v", err)
				tx.Status = "failed"
				txbz, err = cdc.MarshalJSON(&tx)
				txs[i] = txbz
				if err != nil {
					plog.Error().Msgf("Failed to encode failed tx: %v", err)
				}
			}
			tx.Status = "success"
			txbz, err = cdc.MarshalJSON(&tx)
			txs[i] = txbz
			if err != nil {
				plog.Error().Msgf("Failed to encode failed tx: %v", err)
			}
			continue
		}

		if strings.EqualFold(tx.Asset.Network, "Herdius") {

			// Verify if sender has an address for corresponding external asset
			if !strings.EqualFold(tx.Asset.Symbol, "HER") &&
				!isExternalAssetAddressExist(&senderAccount, tx.Asset.Symbol) {
				tx.Status = "failed"
				txbz, err = cdc.MarshalJSON(&tx)
				txs[i] = txbz
				if err != nil {
					plog.Error().Msgf("Failed to encode failed tx: %v", err)
				}
				continue
			}

			// Get Reciever's Account
			rcvrAddressBytes := []byte(tx.RecieverAddress)
			rcvrActbz, _ := stateTrie.TryGet(rcvrAddressBytes)

			var rcvrAccount statedb.Account

			err = cdc.UnmarshalJSON(rcvrActbz, &rcvrAccount)

			if err != nil {
				plog.Error().Msgf("Failed to Unmarshal receiver's account: %v", err)
				continue
			}

			// Verify if Receiver has an address for corresponding external asset
			if !strings.EqualFold(tx.Asset.Symbol, "HER") &&
				!isExternalAssetAddressExist(&rcvrAccount, tx.Asset.Symbol) {
				tx.Status = "failed"
				txbz, err = cdc.MarshalJSON(&tx)
				txs[i] = txbz
				if err != nil {
					plog.Error().Msgf("Failed to encode failed tx: %v", err)
				}
				continue
			}

			// TODO: Deduct Fee from Sender's Account when HER Fee is applied
			//Withdraw fund from Sender Account
			withdraw(&senderAccount, tx.Asset.Symbol, tx.Asset.Value)

			// Credit Reciever's Account
			deposit(&rcvrAccount, tx.Asset.Symbol, tx.Asset.Value)

			senderAccount.Nonce = tx.Asset.Nonce
			updatedSenderAccount, err := cdc.MarshalJSON(senderAccount)

			if err != nil {
				plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
			}

			err = stateTrie.TryUpdate(senderAddressBytes, updatedSenderAccount)
			if err != nil {
				plog.Error().Msgf("Failed to update sender's account in state db: %v", err)
			}

			updatedRcvrAccount, err := cdc.MarshalJSON(rcvrAccount)

			if err != nil {
				plog.Error().Msgf("Failed to Marshal receiver's account: %v", err)
			}

			err = stateTrie.TryUpdate(rcvrAddressBytes, updatedRcvrAccount)
			if err != nil {
				plog.Error().Msgf("Failed to update receiver's account in state db: %v", err)
			}

			// TODO: Fee should be credit to intended recipient
		}

		// Mark the tx as success and
		// add the updated tx to batch that will finally be added to singular block
		tx.Status = "success"
		txbz, err = cdc.MarshalJSON(&tx)
		txs[i] = txbz
		if err != nil {
			plog.Error().Msgf("Failed to encode failed tx: %v", err)
		}
	}

	// State Root
	root, err := stateTrie.Commit(nil)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to commit the state trie: %v.", err))
	}

	// Get Merkle Root Hash of all transactions
	mrh := txs.MerkleHash()

	// Create Singular Block Header
	ts := time.Now().UTC()
	baseHeader := &protobuf.BaseHeader{
		Block_ID:    &protobuf.BlockID{},
		LastBlockID: lastBlock.GetHeader().GetBlock_ID(),
		Height:      lastBlock.Header.Height + 1,
		StateRoot:   root,
		Time: &protobuf.Timestamp{
			Seconds: ts.Unix(),
			Nanos:   ts.UnixNano(),
		},
		RootHash: mrh,
		TotalTxs: uint64(len(txs)),
	}
	blockHashBz, err := cdc.MarshalJSON(baseHeader)
	if err != nil {
		plog.Error().Msgf("Base Header marshaling failed.: %v", err)
	}

	blockHash := hehash.Sum(blockHashBz)
	baseHeader.GetBlock_ID().BlockHash = blockHash
	// Add Header to Block

	s.writerMutex.Lock()
	baseBlock := &protobuf.BaseBlock{
		Header:  baseHeader,
		TxsData: &protobuf.TxsData{Tx: txs},
	}
	s.writerMutex.Unlock()

	// Remove processed transactions from Memory Pool
	mp.RemoveTxs(len(txs))
	return baseBlock, nil
}

func isExternalAssetAddressExist(account *statedb.Account, assetSymbol string) bool {
	if account != nil && account.EBalances != nil &&
		len(account.EBalances[assetSymbol].Address) > 0 {
		return true
	}
	return false
}

func updateAccount(senderAccount *statedb.Account, tx *pluginproto.Tx) *statedb.Account {
	if strings.EqualFold(strings.ToUpper(tx.Asset.Symbol), "HER") &&
		len(senderAccount.Address) == 0 {
		senderAccount.Address = tx.SenderAddress
		senderAccount.Balance = 0
		senderAccount.Nonce = 0
		senderAccount.PublicKey = tx.SenderPubkey
	} else if strings.EqualFold(strings.ToUpper(tx.Asset.Symbol), "HER") &&
		tx.SenderAddress == senderAccount.Address {
		senderAccount.Balance += tx.Asset.Value
		senderAccount.Nonce = tx.Asset.Nonce
	} else if !strings.EqualFold(strings.ToUpper(tx.Asset.Symbol), "HER") &&
		tx.SenderAddress == senderAccount.Address {
		//Register External Asset Addresses
		// Check if an entry already exist for an address
		if eBalance, ok := senderAccount.EBalances[tx.Asset.Symbol]; ok {
			//check if eBalance.Address matches with the asset.address in tx request
			if tx.Asset.ExternalSenderAddress == senderAccount.EBalances[tx.Asset.Symbol].Address {
				if len(senderAccount.EBalances) == 0 {
					eBalance = statedb.EBalance{}
					eBalance.Address = tx.Asset.ExternalSenderAddress
					eBalance.Balance = 0
					eBalance.LastBlockHeight = 0
					eBalance.Nonce = 0
					eBalances := make(map[string]statedb.EBalance)
					eBalances[tx.Asset.Symbol] = eBalance
					senderAccount.EBalances = eBalances
					senderAccount.Nonce = tx.Asset.Nonce
				} else {
					eBalance.Balance += tx.Asset.Value
					if tx.Asset.ExternalBlockHeight > 0 {
						eBalance.LastBlockHeight = tx.Asset.ExternalBlockHeight
					}
					if tx.Asset.ExternalNonce > 0 {
						eBalance.Nonce = tx.Asset.ExternalNonce
					}
					senderAccount.EBalances[tx.Asset.Symbol] = eBalance
					senderAccount.Nonce = tx.Asset.Nonce
				}

			}
		} else {
			eBalance = statedb.EBalance{}
			eBalance.Address = tx.Asset.ExternalSenderAddress
			eBalance.Balance = 0
			eBalance.LastBlockHeight = 0
			eBalance.Nonce = 0
			eBalances := senderAccount.EBalances
			if eBalances == nil || len(eBalances) == 0 {
				eBalances = make(map[string]statedb.EBalance)
			}
			eBalances[tx.Asset.Symbol] = eBalance
			senderAccount.EBalances = eBalances
			senderAccount.Nonce = tx.Asset.Nonce
		}
	}
	return senderAccount
}

// Debit Sender's Account
func withdraw(senderAccount *statedb.Account, assetSymbol string, txValue uint64) {
	if strings.EqualFold(assetSymbol, "HER") {
		balance := senderAccount.Balance
		if balance >= txValue {
			senderAccount.Balance -= txValue
		}
	} else {
		// Get balance of the required external asset
		eBalance := senderAccount.EBalances[strings.ToUpper(assetSymbol)]
		if eBalance.Balance >= txValue {
			eBalance.Balance -= txValue
			senderAccount.EBalances[strings.ToUpper(assetSymbol)] = eBalance
		}
	}
}

// Credit Receiver's Account
func deposit(receiverAccount *statedb.Account, assetSymbol string, txValue uint64) {
	if strings.EqualFold(assetSymbol, "HER") {
		receiverAccount.Balance += txValue
	} else {
		// Get balance of the required external asset
		eBalance := receiverAccount.EBalances[strings.ToUpper(assetSymbol)]
		eBalance.Balance += txValue
		receiverAccount.EBalances[strings.ToUpper(assetSymbol)] = eBalance
	}
}

// LoadStateDBWithInitialAccounts loads state db with initial predefined accounts.
// Initially 50 accounts will be loaded to state db
func LoadStateDBWithInitialAccounts() ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parent := filepath.Dir(wd)
	for i := 0; i < 10; i++ {
		ser := i + 1
		filePath := filepath.Join(parent + "/herdius-core/cmd/testdata/secp205k1Accts/" + strconv.Itoa(ser) + "_peer_id.json")

		nodeKey, err := key.LoadOrGenNodeKey(filePath)

		if err != nil {
			plog.Error().Msgf("Failed to Load or create node keys: %v", err)
		} else {
			pubKey := nodeKey.PrivKey.PubKey()
			b64PubKey := b64.StdEncoding.EncodeToString(pubKey.Bytes())
			//All 10 intital accounts will have an initial balance of 10000 HER tokens
			account := statedb.Account{
				PublicKey: b64PubKey,
				Nonce:     0,
				Address:   pubKey.GetAddress(),
				Balance:   10000,
			}

			actbz, _ := cdc.MarshalJSON(account)
			pubKeyBytes := []byte(pubKey.GetAddress())
			err := trie.TryUpdate(pubKeyBytes, actbz)
			if err != nil {
				plog.Error().Msgf("Failed to store account in state db: %v", err)
			}
		}
	}

	root, err := trie.Commit(nil)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to commit the state trie: %v.", err))
	}

	return root, nil
}

// ShardToValidators distributes a series of childblocks to a series of validators
func (s *Supervisor) ShardToValidators(txs *txbyte.Txs, net *network.Network, stateRoot []byte) error {
	numValds := len(s.Validator)
	if numValds <= 0 {
		return fmt.Errorf("not enough validators in pool to shard, # validators: %v", numValds)
	}
	numTxs := len(*txs)
	numCbs := len(s.ChildBlock)
	var numGrps int

	if numValds <= 3 {
		numGrps = numValds
	} else if numValds%3 == 0 {
		numGrps = numValds % 3
	} else if numValds%2 == 0 {
		numGrps = numValds % 2
	} else {
		numGrps = numValds / 3
	}
	numCbs = numGrps
	fmt.Printf("Number of txs (%v), child blocks (%v), validators (%v)\n", numTxs, numCbs, numValds)

	previousBlockHash := make([]byte, 0)
	txStr := transaction.Tx{}
	txlist := &transaction.TxList{}
	for _, txb := range *txs {
		err := cdc.UnmarshalJSON(txb, &txStr)
		if err != nil {
			return fmt.Errorf("unable to unmarshal tx: %v", err)
		}
		txlist.Transactions = append(txlist.Transactions, &txStr)
	}

	cb := s.CreateChildBlock(net, txlist, 1, previousBlockHash)
	ctx := network.WithSignMessage(context.Background(), true)
	cbmsg := &protobuf.ChildBlockMessage{
		ChildBlock: cb,
	}

	fmt.Println("Broadcasting child block to Validator:", s.Validator[0].Address)
	net.BroadcastByAddresses(ctx, cbmsg, s.Validator[0].Address)
	return nil
}
