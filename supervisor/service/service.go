package service

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	pluginproto "github.com/herdius/herdius-core/hbi/protobuf"

	"github.com/ethereum/go-ethereum/common"
	"github.com/herdius/herdius-core/aws"
	"github.com/herdius/herdius-core/blockchain/protobuf"
	hehash "github.com/herdius/herdius-core/crypto/herhash"
	"github.com/herdius/herdius-core/crypto/merkle"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	cmn "github.com/herdius/herdius-core/libs/common"
	cryptokeys "github.com/herdius/herdius-core/p2p/crypto"
	plog "github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
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
	SetBackup(bool)
	GetChildBlockMerkleHash() ([]byte, error)
	GetValidatorGroupHash() ([]byte, error)
	GetNextValidatorGroupHash() ([]byte, error)
	CreateBaseBlock(lastBlock *protobuf.BaseBlock) (*protobuf.BaseBlock, error)
	GetMutex() *sync.Mutex
	ProcessTxs(lastBlock *protobuf.BaseBlock, net *network.Network) (*protobuf.BaseBlock, error)
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
	Validator           map[string]*protobuf.Validator
	ValidatorChildblock map[string]*protobuf.BlockID //Validator address pointing to child block hash
	VoteInfoData        map[string][]*protobuf.VoteInfo
	stateRoot           []byte
	env                 string
	waitTime            int
	noOfPeersInGroup    int
	backup              bool
}

// StateRoot returns Supervisor current state root
func (s *Supervisor) StateRoot() []byte {
	return s.stateRoot
}

// SetStateRoot sets Supervisor state root
func (s *Supervisor) SetStateRoot(stateRoot []byte) {
	s.stateRoot = stateRoot
}

// Env returns environment name
func (s *Supervisor) Env() string {
	return s.env
}

// SetEnv sets Supervisor environment
func (s *Supervisor) SetEnv(env string) {
	s.env = env
}

// WaitTime returns wait time value
func (s *Supervisor) WaitTime() int {
	return s.waitTime
}

// SetWaitTime sets Supervisor wait time
func (s *Supervisor) SetWaitTime(waitTime int) {
	s.waitTime = waitTime
}

// SetBackup sets Supervisor backup to S3 process
func (s *Supervisor) SetBackup(backup bool) {
	s.backup = backup
}

// Backup sets Supervisor backup to S3 process
func (s *Supervisor) Backup() bool {
	return s.backup
}

// NoOfPeersInGroup ...
func (s *Supervisor) NoOfPeersInGroup() int {
	return s.noOfPeersInGroup
}

// SetNoOfPeersInGroup ...
func (s *Supervisor) SetNoOfPeersInGroup(n int) {
	s.noOfPeersInGroup = n
}

//GetMutex ...
func (s *Supervisor) GetMutex() *sync.Mutex {
	return s.writerMutex
}

// AddValidator adds a validator to group
func (s *Supervisor) AddValidator(publicKey []byte, address string) error {
	if s.Validator == nil {
		s.Validator = make(map[string]*protobuf.Validator)
	}
	validator := &protobuf.Validator{
		Address:      address,
		PubKey:       publicKey,
		Stakingpower: 100,
	}
	s.writerMutex.Lock()
	s.Validator[address] = validator
	s.writerMutex.Unlock()

	return nil
}

// RemoveValidator removes a validator from validators list
func (s *Supervisor) RemoveValidator(address string) {
	s.writerMutex.Lock()
	delete(s.Validator, address)
	s.writerMutex.Unlock()
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
	numValidators := len(s.Validator)
	if numValidators == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("No Child block available: %v.", s.ChildBlock))
	}
	vlBzs := make([][]byte, numValidators)

	for _, validator := range s.Validator {
		vl := validator
		vlBz, err := cdc.MarshalBinaryBare(*vl)
		if err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("Validator Marshaling failed: %v.", err))
		}
		vlBzs = append(vlBzs, vlBz)
	}

	return merkle.SimpleHashFromByteSlices(vlBzs), nil
}

// GetNextValidatorGroupHash ([]byte, error) creates merkle hash of all the next validators
// TODO: refine logic, currently, it's the same as s.GetValidatorGroupHash.
func (s *Supervisor) GetNextValidatorGroupHash() ([]byte, error) {
	return s.GetValidatorGroupHash()
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
		var cbh cmn.HexBytes = v.GetHeader().GetBlockID().GetBlockHash()
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
		StateRoot:              s.stateRoot,
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
func (s *Supervisor) ProcessTxs(lastBlock *protobuf.BaseBlock, net *network.Network) (*protobuf.BaseBlock, error) {
	mp := mempool.GetMemPool()
	select {
	case <-time.After(time.Duration(s.waitTime) * time.Second):
		txs := mp.GetTxs()
		if len(s.Validator) == 0 || len(*txs) == 0 {
			log.Printf("Block creation wait time (%d) elapsed, creating singular base block but with %v transactions", s.waitTime, len(*txs))
			baseBlock, err := s.createSingularBlock(lastBlock, net, *txs, mp, s.stateRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to create singular base block: %v", err)
			}
			mp.RemoveTxs(len(*txs))
			if !s.Backup() {
				log.Println("Backup value false, not backing up block or state")
				return baseBlock, nil
			}
			backuper := aws.NewBackuper(s.env)
			succ, err := backuper.TryBackupBaseBlock(lastBlock, baseBlock)
			if err != nil {
				log.Println("nonfatal: failed to backup to S3:", err)
			} else if !succ {
				log.Println("S3 backup criteria not met; proceeding to backup all unbacked base blocks")
				err := backuper.BackupNeededBaseBlocks(baseBlock)
				if err != nil {
					log.Println("nonfatal: failed to backup both single new and all unbacked base blocks:", err)
				}
				log.Print("Successfully re-evaluated chain and backed up to S3")
			}

			return baseBlock, nil
		}
		err := s.ShardToValidators(txs, net, s.stateRoot)
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
		return nil, fmt.Errorf("failed to retrieve the state trie: %v", err)
	}
	if accountStorage != nil {
		stateTrie = updateStateWithNewExternalBalance(stateTrie)
	}
	if _, err := s.updateStateForTxs(&txs, stateTrie); err != nil {
		return nil, fmt.Errorf("failed to update state for txs: %v", err)
	}

	// Get Merkle Root Hash of all transactions
	mrh := txs.MerkleHash()

	// Create Singular Block Header
	ts := time.Now().UTC()
	baseHeader := &protobuf.BaseHeader{
		Block_ID:    &protobuf.BlockID{},
		LastBlockID: lastBlock.GetHeader().GetBlock_ID(),
		Height:      lastBlock.Header.Height + 1,
		StateRoot:   s.stateRoot,
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

	return baseBlock, nil
}

func updateStateWithNewExternalBalance(stateTrie statedb.Trie) statedb.Trie {
	updateAccs := accountStorage.GetAll()
	log.Println("Total Accounts to update", len(updateAccs))
	for address, item := range updateAccs {
		accountInAccountCache := item
		account := item.Account
		for assetSymbol := range account.EBalances {
			for _, eb := range account.EBalances[assetSymbol] {
				storageKey := assetSymbol + "-" + eb.Address
				IsFirstEntry := item.IsFirstEntry[storageKey]
				IsNewAmountUpdate := item.IsNewAmountUpdate[storageKey]
				if IsNewAmountUpdate && !IsFirstEntry {
					log.Printf("Account from cache to be persisted to state: %v", account)
					sactbz, err := cdc.MarshalJSON(account)
					if err != nil {
						plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
						continue
					}
					stateTrie.TryUpdate([]byte(address), sactbz)
					accountInAccountCache.IsNewAmountUpdate[storageKey] = false
					accountStorage.Set(address, accountInAccountCache)
				}
				if IsFirstEntry {
					log.Println("Account from cache to be persisted to state first time: ", account)
					sactbz, err := cdc.MarshalJSON(account)
					if err != nil {
						plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
						continue
					}
					stateTrie.TryUpdate([]byte(address), sactbz)
					accountInAccountCache.IsFirstEntry[storageKey] = false
					accountStorage.Set(address, accountInAccountCache)
				}
			}
		}

		// IF ERC20Address is presend update accoun balance
		if len(account.Erc20Address) > 0 {
			IsFirstEntry := item.IsFirstHEREntry
			IsNewAmountUpdate := item.IsNewHERAmountUpdate
			if IsNewAmountUpdate && !IsFirstEntry {
				log.Printf("Account from cache to be persisted to state: %v", account)
				sactbz, err := cdc.MarshalJSON(account)
				if err != nil {
					plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
					continue
				}
				stateTrie.TryUpdate([]byte(address), sactbz)
				accountInAccountCache.IsNewHERAmountUpdate = false
				accountStorage.Set(address, accountInAccountCache)
			}
			if IsFirstEntry {
				log.Println("Account from cache to be persisted to state first time: ", account)
				sactbz, err := cdc.MarshalJSON(account)
				if err != nil {
					plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
					continue
				}
				stateTrie.TryUpdate([]byte(address), sactbz)
				accountInAccountCache.IsFirstHEREntry = false
				accountStorage.Set(address, accountInAccountCache)
			}

		}

	}
	return stateTrie
}
func isExternalAssetAddressExist(account *statedb.Account, assetSymbol, assetAddress string) bool {
	if account == nil || account.EBalances == nil {
		return false
	}
	if len(account.EBalances[assetSymbol][assetAddress].Address) > 0 {
		return true
	}
	return false
}

func updateAccountLockedBalance(senderAccount *statedb.Account, tx *pluginproto.Tx) *statedb.Account {
	if senderAccount.LockedBalance == nil {
		senderAccount.LockedBalance = make(map[string]map[string]uint64)
	}
	asset := strings.ToUpper(tx.Asset.Symbol)
	if senderAccount.LockedBalance[asset] == nil {
		senderAccount.LockedBalance[asset] = make(map[string]uint64)
	}

	if tx.SenderAddress == senderAccount.Address {
		senderAccount.LockedBalance[asset][tx.Asset.ExternalSenderAddress] += tx.Asset.LockedAmount
	}
	withdraw(senderAccount, tx.Asset.Symbol, tx.Asset.ExternalSenderAddress, tx.Asset.LockedAmount)
	senderAccount.Nonce = tx.Asset.Nonce
	if strings.EqualFold("BTC", tx.Asset.Symbol) {
		if _, ok := senderAccount.EBalances["HBTC"]; !ok {
			eBalance := statedb.EBalance{}
			eBalance.Address = senderAccount.FirstExternalAddress["ETH"]
			eBalance.Balance = 0
			eBalance.LastBlockHeight = 0
			eBalance.Nonce = 1
			eBalances := senderAccount.EBalances
			eBalances["HBTC"] = make(map[string]statedb.EBalance)
			eBalances["HBTC"][senderAccount.FirstExternalAddress["ETH"]] = eBalance
			senderAccount.EBalances = eBalances
		}
	}
	log.Printf("Locked Account: %v+\n", *senderAccount)
	return senderAccount
}

func updateRedeemAccountLockedBalance(senderAccount *statedb.Account, tx *pluginproto.Tx) *statedb.Account {
	if senderAccount.LockedBalance == nil {
		return senderAccount
	}
	asset := strings.ToUpper(tx.Asset.Symbol)
	if senderAccount.LockedBalance[asset] == nil {
		return senderAccount
	}
	if tx.SenderAddress == senderAccount.Address &&
		tx.Asset.RedeemedAmount <= senderAccount.LockedBalance[asset][tx.Asset.ExternalSenderAddress] {
		senderAccount.LockedBalance[asset][tx.Asset.ExternalSenderAddress] -= tx.Asset.RedeemedAmount
	}
	senderAccount.Nonce = tx.Asset.Nonce
	deposit(senderAccount, tx.Asset.Symbol, tx.Asset.ExternalSenderAddress, tx.Asset.RedeemedAmount)
	log.Printf("Redeemed Account: %v+\n", *senderAccount)
	return senderAccount
}
func updateAccount(senderAccount *statedb.Account, tx *pluginproto.Tx) *statedb.Account {
	if strings.EqualFold(strings.ToUpper(tx.Asset.Symbol), "HER") &&
		len(senderAccount.Address) == 0 {
		senderAccount.Address = tx.SenderAddress
		senderAccount.Balance = 0
		senderAccount.Nonce = 0
		senderAccount.PublicKey = tx.SenderPubkey
		log.Println("Account register", tx)
		senderAccount.Erc20Address = tx.Asset.ExternalSenderAddress
		senderAccount.FirstExternalAddress = make(map[string]string)
	} else if strings.EqualFold(strings.ToUpper(tx.Asset.Symbol), "HER") &&
		tx.SenderAddress == senderAccount.Address {
		senderAccount.Balance += tx.Asset.Value
		senderAccount.Nonce = tx.Asset.Nonce
	} else if !strings.EqualFold(strings.ToUpper(tx.Asset.Symbol), "HER") &&
		tx.SenderAddress == senderAccount.Address {

		// Update account's Nonce
		senderAccount.Nonce = tx.Asset.Nonce

		// Register External Asset Addresses if not exist
		if assetEBalance, ok := senderAccount.EBalances[tx.Asset.Symbol]; ok {
			if _, ok := assetEBalance[tx.Asset.ExternalSenderAddress]; !ok {
				eBalance := statedb.EBalance{}
				eBalance.Address = tx.Asset.ExternalSenderAddress
				eBalance.Balance = 0
				eBalance.LastBlockHeight = 0
				eBalance.Nonce = 0
				eBalances := senderAccount.EBalances
				eBalances[tx.Asset.Symbol][tx.Asset.ExternalSenderAddress] = eBalance
				senderAccount.EBalances = eBalances
			}
		} else {
			eBalance := statedb.EBalance{}
			eBalance.Address = tx.Asset.ExternalSenderAddress
			eBalance.Balance = 0
			eBalance.LastBlockHeight = 0
			eBalance.Nonce = 0
			eBalances := senderAccount.EBalances
			if len(eBalances) == 0 {
				eBalances = make(map[string]map[string]statedb.EBalance)
			}
			if len(eBalances[tx.Asset.Symbol]) == 0 {
				eBalances[tx.Asset.Symbol] = make(map[string]statedb.EBalance)
			}
			eBalances[tx.Asset.Symbol][tx.Asset.ExternalSenderAddress] = eBalance
			senderAccount.EBalances = eBalances
			if senderAccount.FirstExternalAddress == nil {
				senderAccount.FirstExternalAddress = make(map[string]string)
			}
			senderAccount.FirstExternalAddress[tx.Asset.Symbol] = tx.Asset.ExternalSenderAddress
		}
	}
	return senderAccount
}

// Debit Sender's Account
func withdraw(senderAccount *statedb.Account, assetSymbol, assetExtAddress string, txValue uint64) {
	if strings.EqualFold(assetSymbol, "HER") {
		balance := senderAccount.Balance
		if balance >= txValue {
			senderAccount.Balance -= txValue
		}
	} else {
		// Get balance of the required external asset
		eBalance := senderAccount.EBalances[strings.ToUpper(assetSymbol)][assetExtAddress]
		if eBalance.Balance >= txValue {
			eBalance.Balance -= txValue
			senderAccount.EBalances[strings.ToUpper(assetSymbol)][assetExtAddress] = eBalance
		}
	}
}

// Credit Receiver's Account
func deposit(receiverAccount *statedb.Account, assetSymbol, assetExtAddress string, txValue uint64) {
	if strings.EqualFold(assetSymbol, "HER") {
		receiverAccount.Balance += txValue
	} else {
		// Get balance of the required external asset
		eBalance := receiverAccount.EBalances[strings.ToUpper(assetSymbol)][assetExtAddress]
		eBalance.Balance += txValue
		receiverAccount.EBalances[strings.ToUpper(assetSymbol)][assetExtAddress] = eBalance
	}
}

func (s *Supervisor) validatorAddresses() []string {
	addresses := make([]string, len(s.Validator))
	for address := range s.Validator {
		addresses = append(addresses, address)
	}
	return addresses
}

func (s *Supervisor) validatorGroups(numGroup int) [][]string {
	s.writerMutex.Lock()
	defer s.writerMutex.Unlock()

	numValds := len(s.Validator)
	validators := make([]string, 0)
	for address := range s.Validator {
		validators = append(validators, address)
	}

	var groups [][]string
	if numGroup <= 0 {
		numGroup = 1
	}
	groupSize := (numValds + numGroup - 1) / numGroup

	for start := 0; start < numValds; start += groupSize {
		end := start + groupSize
		if end > numValds {
			end = numValds
		}
		groups = append(groups, validators[start:end])
	}

	return groups
}

func (s *Supervisor) txsGroups(txList *transaction.TxList, numGroup int) [][]*transaction.Tx {
	txs := txList.Transactions

	var groups [][]*transaction.Tx
	groupSize := (len(txs) + numGroup - 1) / numGroup

	for start := 0; start < len(txs); start += groupSize {
		end := start + groupSize
		if end > len(txs) {
			end = len(txs)
		}
		groups = append(groups, txs[start:end])
	}

	return groups
}

// ShardToValidators distributes a series of childblocks to a series of validators
func (s *Supervisor) ShardToValidators(txs *txbyte.Txs, net *network.Network, stateRoot []byte) error {
	numValds := len(s.Validator)
	if numValds <= 0 {
		return fmt.Errorf("not enough validators in pool to shard, # validators: %v", numValds)
	}
	numTxs := len(*txs)
	// TODO: Make number of groups configurable.
	numGroup := 5
	vGroups := s.validatorGroups(numGroup)
	fmt.Printf("Number of txs (%v), groups (%v), validators (%v)\n", numTxs, vGroups, numValds)

	if len(stateRoot) == 0 {
		return fmt.Errorf("cannot process an empty stateRoot for the trie")
	}
	stateTrie, err := statedb.NewTrie(common.BytesToHash(stateRoot))
	if err != nil {
		return fmt.Errorf("error attempting to retrieve state db trie from stateRoot: %v", err)
	}
	if accountStorage != nil {
		stateTrie = updateStateWithNewExternalBalance(stateTrie)
	}
	txList, err := s.updateStateForTxs(txs, stateTrie)
	if err != nil {
		return fmt.Errorf("failed to update state for txs: %v", err)
	}

	txsGroups := s.txsGroups(txList, numGroup)
	previousBlockHash := make([]byte, 0)

	for i := range txsGroups {
		cb := s.CreateChildBlock(net, &transaction.TxList{Transactions: txsGroups[i]}, int64(len(txsGroups[i])), previousBlockHash)
		previousBlockHash = cb.GetHeader().GetBlockID().BlockHash
		if i > 0 {
			cb.GetHeader().GetLastBlockID().BlockHash = previousBlockHash
		}
		ctx := network.WithSignMessage(context.Background(), true)
		cbmsg := &protobuf.ChildBlockMessage{ChildBlock: cb}
		fmt.Println("Broadcasting child block to Validator Group:", vGroups[i])
		net.BroadcastByAddresses(ctx, cbmsg, vGroups[i]...)
	}

	return nil
}

func (s *Supervisor) updateStateForTxs(txs *txbyte.Txs, stateTrie statedb.Trie) (*transaction.TxList, error) {
	txStr := transaction.Tx{}
	txlist := &transaction.TxList{}
	tx := pluginproto.Tx{}
	for i, txbz := range *txs {
		err := cdc.UnmarshalJSON(txbz, &txStr)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal tx: %v", err)
		}
		txlist.Transactions = append(txlist.Transactions, &txStr)

		err = cdc.UnmarshalJSON(txbz, &tx)
		if err != nil {
			log.Printf("Failed to Unmarshal tx: %v", err)
			continue
		}

		// Get the public key of the sender
		senderAddress := tx.GetSenderAddress()
		pubKeyS, err := b64.StdEncoding.DecodeString(tx.GetSenderPubkey())
		if err != nil {
			log.Printf("Failed to decode sender public key: %v", err)
			plog.Error().Msgf("Failed to decode sender public key: %v", err)
			tx.Status = "failed"
			txbz, err = cdc.MarshalJSON(&tx)
			(*txs)[i] = txbz
			if err != nil {
				log.Printf("Failed to encode failed tx: %v", err)
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
			LockedAmount:          tx.Asset.LockedAmount,
			RedeemedAmount:        tx.Asset.RedeemedAmount,
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
			log.Printf("Failed to marshal the transaction to verify sign: %v", err)
			continue
		}

		decodedSig, err := b64.StdEncoding.DecodeString(tx.Sign)

		if err != nil {
			plog.Error().Msgf("Failed to decode the base64 sign to verify sign: %v", err)
			log.Printf("Failed to decode the base64 sign to verify sign: %v", err)
			continue
		}

		signVerificationRes := pubKey.VerifyBytes(txbBeforeSign, decodedSig)
		if !signVerificationRes {
			plog.Error().Msgf("Signature Verification Failed: %v", signVerificationRes)
			log.Printf("Signature Verification Failed: %v", signVerificationRes)
			tx.Status = "failed"
			txbz, err = cdc.MarshalJSON(&tx)
			(*txs)[i] = txbz
			if err != nil {
				plog.Error().Msgf("Failed to encode failed tx: %v", err)
				log.Printf("Failed to encode failed tx: %v", err)
			}
			continue
		}
		var senderAccount statedb.Account
		senderAddressBytes := []byte(senderAddress)

		// Get account details from state trie
		senderActbz, err := stateTrie.TryGet(senderAddressBytes)
		if err != nil {
			plog.Error().Msgf("Failed to retrieve account detail: %v", err)
			log.Printf("Failed to retrieve account detail: %v", err)
			continue
		}

		if len(senderActbz) > 0 {
			err = cdc.UnmarshalJSON(senderActbz, &senderAccount)
			if err != nil {
				log.Printf("Failed to Unmarshal account: %v", err)
				plog.Error().Msgf("Failed to Unmarshal account: %v", err)
				continue
			}
		}

		// Check if tx is of type account update
		if strings.EqualFold(tx.Type, "External") {
			symbol := tx.Asset.Symbol
			if symbol != "BTC" && symbol != "ETH" {
				log.Printf("Unsupported external asset symbol: %v", symbol)
				plog.Error().Msgf("Unsupported external asset symbol: %v", symbol)
				continue
			}

			// By default each of the new accounts will have HER token (with balance 0)
			// added to the map object balances
			balance := senderAccount.EBalances[symbol][tx.Asset.ExternalSenderAddress]
			if balance == (statedb.EBalance{}) {
				plog.Error().Msgf("Sender has no assets for the given symbol: %v", symbol)
				log.Printf("Sender has no assets for the given symbol: %v", symbol)
				continue
			}
			if balance.Balance <= tx.Asset.Value {
				plog.Error().Msgf("Sender does not have enough assets in account (%d) to send transaction amount (%d)", balance.Balance, tx.Asset.Value)
				log.Printf("Sender does not have enough assets in account (%d) to send transaction amount (%d)", balance.Balance, tx.Asset.Value)
				continue
			}
			balance.Balance -= tx.Asset.Value

			senderAccount.Nonce = tx.Asset.Nonce
			senderAccount.EBalances[symbol][tx.Asset.ExternalSenderAddress] = balance

			sactbz, err := cdc.MarshalJSON(senderAccount)
			if err != nil {
				log.Printf("Failed to Marshal sender's account: %v", err)
				plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
				continue
			}
			addressBytes := []byte(pubKey.GetAddress())
			err = stateTrie.TryUpdate(addressBytes, sactbz)
			if err != nil {
				log.Printf("Failed to store account in state db: %v", err)
				plog.Error().Msgf("Failed to store account in state db: %v", err)
				tx.Status = "failed"
				txbz, err = cdc.MarshalJSON(&tx)
				(*txs)[i] = txbz
				if err != nil {
					log.Printf("Failed to encode failed tx: %v", err)
					plog.Error().Msgf("Failed to encode failed tx: %v", err)
				}
			}
			tx.Status = "success"
			txbz, err = cdc.MarshalJSON(&tx)
			(*txs)[i] = txbz
			if err != nil {
				log.Printf("Failed to encode failed tx: %v", err)
				plog.Error().Msgf("Failed to encode failed tx: %v", err)
			}

			continue

		} else if strings.EqualFold(tx.Type, "Update") ||
			strings.EqualFold(tx.Type, "Lock") ||
			strings.EqualFold(tx.Type, "Redeem") {

			switch txType := strings.ToUpper(tx.Type); txType {
			case "UPDATE":
				senderAccount = *(updateAccount(&senderAccount, &tx))
			case "LOCK":
				senderAccount = *(updateAccountLockedBalance(&senderAccount, &tx))
			case "REDEEM":
				senderAccount = *(updateRedeemAccountLockedBalance(&senderAccount, &tx))
			}

			sactbz, err := cdc.MarshalJSON(senderAccount)
			if err != nil {
				log.Printf("Failed to Marshal sender's account: %v", err)
				plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
				continue
			}
			addressBytes := []byte(pubKey.GetAddress())
			err = stateTrie.TryUpdate(addressBytes, sactbz)
			if err != nil {
				log.Printf("Failed to store account in state db: %v", err)
				plog.Error().Msgf("Failed to store account in state db: %v", err)
				tx.Status = "failed"
				txbz, err = cdc.MarshalJSON(&tx)
				(*txs)[i] = txbz
				if err != nil {
					log.Printf("Failed to encode failed tx: %v", err)
					plog.Error().Msgf("Failed to encode failed tx: %v", err)
				}
			}
			tx.Status = "success"
			txbz, err = cdc.MarshalJSON(&tx)
			(*txs)[i] = txbz
			if err != nil {
				log.Printf("Failed to encode failed tx: %v", err)
				plog.Error().Msgf("Failed to encode failed tx: %v", err)
			}
			continue
		}

		if strings.EqualFold(tx.Asset.Network, "Herdius") {

			// Verify if sender has an address for corresponding external asset
			if !strings.EqualFold(tx.Asset.Symbol, "HER") &&
				!isExternalAssetAddressExist(&senderAccount, tx.Asset.Symbol, tx.Asset.ExternalSenderAddress) {
				tx.Status = "failed"
				txbz, err = cdc.MarshalJSON(&tx)
				(*txs)[i] = txbz
				if err != nil {
					log.Printf("Failed to encode failed tx: %v", err)
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
				log.Printf("Failed to Unmarshal receiver's account: %v", err)
				plog.Error().Msgf("Failed to Unmarshal receiver's account: %v", err)
				continue
			}

			// Verify if Receiver has an address for corresponding external asset
			if !strings.EqualFold(tx.Asset.Symbol, "HER") && len(rcvrAccount.EBalances[tx.Asset.Symbol]) == 0 {
				tx.Status = "failed"
				txbz, err = cdc.MarshalJSON(&tx)
				(*txs)[i] = txbz
				if err != nil {
					log.Printf("Failed to encode failed tx: %v", err)
					plog.Error().Msgf("Failed to encode failed tx: %v", err)
				}
				continue
			}

			// TODO: Deduct Fee from Sender's Account when HER Fee is applied
			//Withdraw fund from Sender Account
			withdraw(&senderAccount, tx.Asset.Symbol, tx.Asset.ExternalSenderAddress, tx.Asset.Value)

			// Credit Reciever's Account
			// If credit to external address, pick first account
			// TODO: Should we consider tx.Asset.ExternalRecieverAddress?
			deposit(&rcvrAccount, tx.Asset.Symbol, rcvrAccount.FirstExternalAddress[tx.Asset.Symbol], tx.Asset.Value)

			senderAccount.Nonce = tx.Asset.Nonce
			updatedSenderAccount, err := cdc.MarshalJSON(senderAccount)
			if err != nil {
				log.Printf("Failed to Marshal sender's account: %v", err)
				plog.Error().Msgf("Failed to Marshal sender's account: %v", err)
			}

			err = stateTrie.TryUpdate(senderAddressBytes, updatedSenderAccount)
			if err != nil {
				log.Printf("Failed to update sender's account in state db: %v", err)
				plog.Error().Msgf("Failed to update sender's account in state db: %v", err)
			}

			updatedRcvrAccount, err := cdc.MarshalJSON(rcvrAccount)

			if err != nil {
				log.Printf("Failed to Marshal receiver's account: %v", err)
				plog.Error().Msgf("Failed to Marshal receiver's account: %v", err)
			}

			err = stateTrie.TryUpdate(rcvrAddressBytes, updatedRcvrAccount)
			if err != nil {
				log.Printf("Failed to update receiver's account in state db: %v", err)
				plog.Error().Msgf("Failed to update receiver's account in state db: %v", err)
			}

			// TODO: Fee should be credit to intended recipient
		}

		// Mark the tx as success and
		// add the updated tx to batch that will finally be added to singular block
		tx.Status = "success"
		txbz, err = cdc.MarshalJSON(&tx)
		(*txs)[i] = txbz
		if err != nil {
			log.Printf("Failed to encode failed tx: %v", err)
			plog.Error().Msgf("Failed to encode failed tx: %v", err)
		}

	}

	root, err := stateTrie.Commit(nil)
	if err != nil {
		log.Println("Failed to commit to state trie:", err)
		plog.Error().Msgf("Failed to commit to state trie: %v", err)
	}
	s.SetStateRoot(root)
	return txlist, nil
}
