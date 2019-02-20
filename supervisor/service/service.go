package service

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/crypto"
	hehash "github.com/herdius/herdius-core/crypto/herhash"
	"github.com/herdius/herdius-core/crypto/merkle"
	cmn "github.com/herdius/herdius-core/libs/common"
	cryptokeys "github.com/herdius/herdius-core/p2p/crypto"
	"github.com/herdius/herdius-core/p2p/key"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"

	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/herdius/herdius-core/supervisor/transaction"
	txbyte "github.com/herdius/herdius-core/tx"
)

// SupervisorI is an interface
type SupervisorI interface {
	AddValidator(publicKey []byte, address string) error
	RemoveValidator(address string)
	CreateChildBlock(net *network.Network, txs *transaction.TxList, height int64, previousBlockHash []byte) *protobuf.ChildBlock
	CreateTxBatchesFromFile(filename string, numOfBatches, txsNum int, stateRoot []byte) error
	CreateTxBatches()
	SetWriteMutex()
	GetChildBlockMerkleHash() ([]byte, error)
	GetValidatorGroupHash() ([]byte, error)
	GetNextValidatorGroupHash() ([]byte, error)
	CreateBaseBlock(lastBlock *protobuf.BaseBlock) (*protobuf.BaseBlock, error)
	GetMutex() *sync.Mutex
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
	return nil, fmt.Errorf(fmt.Sprintf("No Child block available: %v.", s.ChildBlock))
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

// CreateTxBatchesFromFile loads transactions from a local file.
func (s *Supervisor) CreateTxBatchesFromFile(filename string, numOfBatches, txsNum int, stateRoot []byte) error {

	fileReader, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	// Get Trie Root of state db from last block

	stateTrie, err := ethtrie.New(common.BytesToHash(stateRoot), statedb.GetDB())

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to retrieve the state trie: %v.", err))
	}

	dec := json.NewDecoder(fileReader)
	var txService transaction.Service
	txService = transaction.TxService()

	for {
		var tx transaction.Tx
		err := dec.Decode(&tx)
		if err == io.EOF {
			// all txs are done
			break
		}
		if err != nil {
			log.Error().Msgf("Failed while reading the file: %v", err)
		}

		// Verify the account state from Tx and update it in state trie
		var pubKey crypto.PubKey

		err = cdc.UnmarshalBinaryBare(tx.Senderpubkey, &pubKey)
		if err != nil {
			log.Error().Msgf("Failed to Unmarshal senderkey: %v", err)
		}

		actbz, _ := stateTrie.TryGet(pubKey.Bytes())

		var account statedb.Account
		err = cdc.UnmarshalJSON(actbz, &account)

		if err != nil {
			log.Error().Msgf("Failed to Unmarshal the account from state trie: %v", err)
		}

		// Increment the Account nonce by 1
		account.Nonce = account.Nonce + 1
		txFee, _ := strconv.Atoi(string(tx.Fee))

		bal, _ := strconv.Atoi(string(account.Balance))

		//Check if tx fee is less than available balance otherwise proceed to next tx
		if int(txFee) > bal {
			continue
		}

		//Deduct transaction fee from the available balance in the account
		bal = bal - int(txFee)
		account.Balance = []byte(strconv.Itoa(bal))

		actbz, _ = cdc.MarshalJSON(account)
		err = stateTrie.TryUpdate(pubKey.Bytes(), actbz)
		if err != nil {
			log.Error().Msgf("Failed to update the account to state trie: %v", err)
		}
		txService.AddTx(tx)

	}

	newStateRoot, err := stateTrie.Commit(nil)

	s.StateRoot = newStateRoot.Bytes()

	txList := txService.GetTxList()
	var txNum int
	if txList == nil {
		return nil
	}

	//Number of transactions loaded from the json file
	txNum = len((*txList).Transactions)

	if txNum == 0 {

		return nil
	}

	allTxs := (*txList).Transactions
	var txCounter int
	txCounter = 0
	txBatches := make([]txbyte.Txs, 0)

	for b := 1; b <= numOfBatches; b++ {
		txs := make([][]byte, txsNum)
		for i := 0; i < txsNum; i++ {
			tx := *allTxs[txCounter]

			txbz, err := cdc.MarshalJSON(tx)

			if err != nil {
				log.Error().Msgf("Failed to Marshal the Tx: %v", err)
			}
			txs[i] = txbz
			txCounter++
		}
		txBatches = append(txBatches, txs)

	}

	s.TxBatches = &txBatches
	return nil
}

// CreateTxBatches creates batches of the transactions
func (s *Supervisor) CreateTxBatches() {}

// CreateBaseBlock creates the base block with all the child blocks
func (s *Supervisor) CreateBaseBlock(lastBlock *protobuf.BaseBlock) (*protobuf.BaseBlock, error) {

	// Create the merkle hash of all the child blocks
	cbMerkleHash, err := s.GetChildBlockMerkleHash()

	if err != nil {
		log.Error().Msgf("Failed to create Merkle Hash of Validators: %v", err)
	}

	// Create the merkle hash of all the validators
	vgHash, err := s.GetValidatorGroupHash()

	if err != nil {
		log.Error().Msgf("Failed to create Merkle Hash of Validators: %v", err)
	}

	// Create the merkle hash of all the next validators
	nvgHash, err := s.GetNextValidatorGroupHash()

	if err != nil {
		log.Error().Msgf("Failed to create Merkle Hash of Next Validators: %v", err)
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
		log.Error().Msgf("Vote commits marshaling failed.: %v", err)
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
		log.Error().Msgf("Base Header marshaling failed.: %v", err)
	}
	blockHash := hehash.Sum(blockHashBz)

	baseHeader.GetBlock_ID().BlockHash = blockHash

	childBlocksBz, err := cdc.MarshalJSON(s.ChildBlock)

	if err != nil {
		log.Error().Msgf("Child blocks marshaling failed.: %v", err)
	}

	// Vote commits marshalling
	vcBz, err := cdc.MarshalJSON(votecommits)
	if err != nil {
		log.Error().Msgf("Vote Commits marshaling failed.: %v", err)
	}

	// Validators marshaling

	valsBz, err := cdc.MarshalJSON(s.Validator)

	if err != nil {
		log.Error().Msgf("Validators marshaling failed.: %v", err)
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
	txservice := txbyte.GetTxService()

	for _, tx := range txList.Transactions {

		txbz, err := cdc.MarshalJSON(*tx)
		//log.Fatalf("Marshalling failed: %v.", err)
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
		address = keys.PubKey.Address().String()
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

// LoadStateDBWithInitialAccounts loads state db with initial predefined accounts.
// Initially 50 accounts will be loaded to state db
func LoadStateDBWithInitialAccounts() ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parent := filepath.Dir(wd)
	for i := 0; i < 50; i++ {
		ser := i + 1
		filePath := filepath.Join(parent + "/herdius-core/supervisor/testdata/accounts/account_" + strconv.Itoa(ser) + ".json")

		nodeKey, err := key.LoadOrGenNodeKey(filePath)

		if err != nil {
			log.Error().Msgf("Failed to Load or create node keys: %v", err)
		} else {
			pubKey := nodeKey.PrivKey.PubKey()

			//All 50 intital accounts will have an initial balance of 10000 HER tokens
			var bal = "10000"
			account := statedb.Account{
				Nonce:       0,
				Address:     pubKey.Address().String(),
				AddressHash: pubKey.Bytes(),
				Balance:     []byte(bal),
			}

			actbz, _ := cdc.MarshalJSON(account)
			err := trie.TryUpdate(pubKey.Bytes(), actbz)
			if err != nil {
				log.Error().Msgf("Failed to store account in state db: %v", err)
			}
		}
	}

	root, err := trie.Commit(nil)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to commit the state trie: %v.", err))
	}
	return root, nil
}
