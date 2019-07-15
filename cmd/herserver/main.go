package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"strconv"

	nlog "log"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/herdius/herdius-core/aws/restore"
	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/blockchain/protobuf"
	blockProtobuf "github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/config"
	cryptokey "github.com/herdius/herdius-core/crypto"
	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/hbi/message"
	protoplugin "github.com/herdius/herdius-core/hbi/protobuf"
	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/p2p/crypto"
	keystore "github.com/herdius/herdius-core/p2p/key"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
	"github.com/herdius/herdius-core/p2p/network/discovery"
	"github.com/herdius/herdius-core/p2p/types/opcode"
	external "github.com/herdius/herdius-core/storage/exbalance"
	syncer "github.com/herdius/herdius-core/syncer"

	"github.com/herdius/herdius-core/storage/state/statedb"

	sup "github.com/herdius/herdius-core/supervisor/service"
	validator "github.com/herdius/herdius-core/validator/service"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()
var supsvc *sup.Supervisor
var blockchainSvc *blockchain.Service
var voteCount = 0

// Flag to check if a new child block has arrived to validator
var isChildBlockReceivedByValidator = false

// Child block message object received
var mcb = &blockProtobuf.ChildBlockMessage{}

// firstPingFromValidator checks whether a connection is established betweer supervisor and validator.
// And it is used to send a message on established connection.
var firstPingFromValidator = 0
var nodeKeydir = "./cmd/testdata/secp205k1Accts/"
var t1 time.Time
var t2 time.Time
var addresses = make([]string, 0)

var accountStorage external.BalanceStorage

// HerdiusMessagePlugin will receive all trasnmitted messages.
type HerdiusMessagePlugin struct{ *network.Plugin }

func init() {

	nlog.SetFlags(nlog.LstdFlags | nlog.Lshortfile)
	supsvc = &sup.Supervisor{}
	supsvc.SetWriteMutex()
	supsvc.ValidatorChildblock = make(map[string]*blockProtobuf.BlockID, 0)
	supsvc.ChildBlock = make([]*blockProtobuf.ChildBlock, 0)
	supsvc.VoteInfoData = make(map[string][]*blockProtobuf.VoteInfo, 0)

	RegisterAminoService(cdc)
}

//RegisterAminoService registers Amino service for message (en/de) coding
func RegisterAminoService(cdc *amino.Codec) {
	cryptoAmino.RegisterAmino(cdc)

}

// Receive handles each received message for both Supervisor and Validator
func (state *HerdiusMessagePlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {
	case *blockProtobuf.ConnectionMessage:
		address := ctx.Client().ID.Address
		pubKey := ctx.Client().ID.PublicKey
		err := supsvc.AddValidator(pubKey, address)
		if err != nil {
			log.Info().Msgf("<%s> Failed to add validator: %v", address, err)
		}

		addresses = append(addresses, address)
		log.Info().Msgf("<%s> %s", address, msg.Message)

		// This map will be used to map validators to their respective child blocks
		mx := supsvc.GetMutex()
		mx.Lock()
		supsvc.ValidatorChildblock[address] = &blockProtobuf.BlockID{}
		mx.Unlock()

		sender, err := ctx.Network().Client(ctx.Client().Address)
		if err != nil {
			return fmt.Errorf("failed to get client network: %v", err)
		}
		nonce := 1
		err = sender.Reply(network.WithSignMessage(context.Background(), true), uint64(nonce),
			&blockProtobuf.ConnectionMessage{Message: "Connection established with Supervisor"})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
		}
	case *blockProtobuf.ChildBlockMessage:
		mcb = msg
		vote := mcb.GetVote()
		if vote != nil {
			// Increment the vote count of validator group
			voteCount++

			var cbhash cmn.HexBytes
			cbhash = mcb.GetChildBlock().GetHeader().GetBlockID().GetBlockHash()
			voteinfo := supsvc.VoteInfoData[cbhash.String()]
			voteinfo = append(voteinfo, vote)
			supsvc.VoteInfoData[cbhash.String()] = voteinfo

			sign := vote.GetSignature()
			var pubKey cryptokey.PubKey

			cdc.UnmarshalBinaryBare(vote.GetValidator().GetPubKey(), &pubKey)

			isVerified := pubKey.VerifyBytes(vote.GetValidator().GetPubKey(), sign)

			isChildBlockSigned := mcb.GetVote().GetSignedCurrentBlock()

			// Check whether Childblock is verified and signed by the validator
			if isChildBlockSigned && isVerified {

				address := ctx.Client().ID.Address
				mx := supsvc.GetMutex()
				mx.Lock()
				supsvc.ValidatorChildblock[address] = mcb.GetChildBlock().GetHeader().GetBlockID()
				mx.Unlock()
				log.Info().Msgf("<%s> Validator verified and signed the child block: %v", address, isVerified)

				// TODO: It needs to be implemented in a proper way
				// It should probably be a part of the consensus on child block
				// How can we do that?
				if voteCount == len(supsvc.Validator) {

					lastBlock := blockchainSvc.GetLastBlock()

					var stateRoot cmn.HexBytes
					stateRoot = supsvc.StateRoot()
					stateTrie, err := statedb.NewTrie(common.BytesToHash(stateRoot))
					if err != nil {
						log.Error().Msgf("Failed to create new state trie: %v", err)
					}
					// TODO: handle err?
					newStateRoot, err := stateTrie.Commit(nil)
					supsvc.SetStateRoot(newStateRoot)
					supsvc.ChildBlock = append(supsvc.ChildBlock, mcb.GetChildBlock())
					// TODO: handle err?
					baseBlock, err := supsvc.CreateBaseBlock(lastBlock)

					err = blockchainSvc.AddBaseBlock(baseBlock)
					if err != nil {
						log.Error().Msgf("Failed to Add Base Block: %v", err)
					}

					var bbh cmn.HexBytes
					bbh = baseBlock.GetHeader().GetBlock_ID().GetBlockHash()
					log.Info().Msg("New Block Added")
					log.Info().Msgf("Block Id: %v", bbh.String())

					log.Info().Msgf("Block Height: %v", baseBlock.GetHeader().GetHeight())

					s := lastBlock.GetHeader().GetTime().GetSeconds()
					ts := time.Unix(s, 0)
					log.Info().Msgf("Timestamp : %v", ts)

					stateRoot = supsvc.StateRoot()
					log.Info().Msgf("State root : %v", stateRoot)
					// Once new base block is added to be block chain
					// do the following

					supsvc.ValidatorChildblock = make(map[string]*blockProtobuf.BlockID, 0)
					supsvc.ChildBlock = make([]*blockProtobuf.ChildBlock, 0)
					supsvc.VoteInfoData = make(map[string][]*blockProtobuf.VoteInfo, 0)
					mcb = &blockProtobuf.ChildBlockMessage{}
					voteCount = 0

					supsvc.SetStateRoot([]byte{0})
					t2 = time.Now()

					diff := t2.Sub(t1)

					log.Info().Msgf("Total time : %v", diff)

				} else {
					log.Info().Msgf("Vote count mismatch, votes (%d) of validators (%d)", voteCount, len(supsvc.Validator))
				}
			} else {
				log.Info().Msgf("<%s> Validator verification or signature verification failed: %v", ctx.Client().ID.Address, isVerified)
			}

		} else {
			isChildBlockReceivedByValidator = true
		}
	}
	return nil
}

func main() {

	// process other flags
	peersFlag := flag.String("peers", "", "peers to connect to")
	supervisorFlag := flag.Bool("supervisor", false, "run as supervisor")
	groupSizeFlag := flag.Int("groupsize", 3, "# of peers in a validator group")
	portFlag := flag.Int("port", 0, "port to bind validator to")
	envFlag := flag.String("env", "dev", "environment to build network and run process for")
	waitTimeFlag := flag.Int("waitTime", 15, "time to wait before the Memory Pool is flushed to a new block")
	restoreFlag := flag.Bool("restore", false, "restore blockchain from S3")
	backupFlag := flag.Bool("backup", false, "backup blockchain to S3")
	flag.Parse()

	noOfPeersInGroup := *groupSizeFlag
	port := *portFlag
	env := *envFlag
	waitTime := *waitTimeFlag
	restr := *restoreFlag
	backup := *backupFlag
	confg := config.GetConfiguration(env)
	peers := []string{}
	if len(*peersFlag) > 0 {
		peers = strings.Split(*peersFlag, ",")
	}

	if port == 0 {
		port = confg.SelfBroadcastPort
	}

	// Generate or Load Keys
	nodeAddress := confg.SelfBroadcastIP + "_" + strconv.Itoa(port)
	nodekey, err := keystore.LoadOrGenNodeKey(nodeKeydir + nodeAddress + "_sk_peer_id.json")
	if err != nil {
		log.Error().Msgf("Failed to create or load node key: %v", err)
	}
	privKey := nodekey.PrivKey
	pubKey := privKey.PubKey()
	keys := &crypto.KeyPair{
		PublicKey:  pubKey.Bytes(),
		PrivateKey: privKey.Bytes(),
		PrivKey:    privKey,
		PubKey:     pubKey,
	}

	opcode.RegisterMessageType(opcode.Opcode(1111), &blockProtobuf.ChildBlockMessage{})
	opcode.RegisterMessageType(opcode.Opcode(1112), &blockProtobuf.ConnectionMessage{})
	opcode.RegisterMessageType(opcode.Opcode(1113), &protoplugin.BlockHeightRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1114), &protoplugin.BlockResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1115), &protoplugin.AccountRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1116), &protoplugin.AccountResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1117), &protoplugin.TxRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1118), &protoplugin.TxResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1119), &protoplugin.TxDetailRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1120), &protoplugin.TxDetailResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1121), &protoplugin.TxsByAddressRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1122), &protoplugin.TxsResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1123), &protoplugin.TxsByAssetAndAddressRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1124), &protoplugin.TxUpdateRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1125), &protoplugin.TxUpdateResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1126), &protoplugin.TxDeleteRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1127), &protoplugin.TxLockedRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1128), &protoplugin.TxLockedResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1129), &protobuf.Ping{})
	opcode.RegisterMessageType(opcode.Opcode(1130), &protobuf.Pong{})
	opcode.RegisterMessageType(opcode.Opcode(1131), &protoplugin.TxRedeemRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1132), &protoplugin.TxRedeemResponse{})

	builder := network.NewBuilder(env)
	builder.SetKeys(keys)

	builder.SetAddress(network.FormatAddress(confg.Protocol, confg.SelfBroadcastIP, uint16(port)))

	// Register peer discovery plugin.
	builder.AddPlugin(new(discovery.Plugin))

	// Add custom Herdius plugin.
	builder.AddPlugin(new(HerdiusMessagePlugin))
	builder.AddPlugin(new(message.BlockMessagePlugin))
	builder.AddPlugin(new(message.AccountMessagePlugin))
	builder.AddPlugin(new(message.TransactionMessagePlugin))

	net, err := builder.Build()
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	go net.Listen()
	defer net.Close()

	c := new(network.ConnTester)
	go func() {
		c.IsConnected(net, peers)
	}()

	// As of now Databases will only be loaded for Supervisor.
	// Chain data and state information will be stored at supervisor's node.

	var stateRoot []byte
	if *supervisorFlag {
		accountStorage = external.New()
		if restr {
			nlog.Println("Restore value true: proceeding to restore from AWS S3")
			r := restore.NewRestorer(env, 3)
			err := r.Restore()
			if err != nil {
				log.Error().Msg(err.Error())
			}
		}
		blockchain.LoadDB()
		sup.LoadStateDB(accountStorage)
		blockchainSvc := &blockchain.Service{}

		lastBlock := blockchainSvc.GetLastBlock()

		go syncer.SyncAllAccounts(accountStorage)

		if err != nil {
			log.Error().Msgf("Failed while getting last block: %v\n", err)
		} else {
			var lbh cmn.HexBytes
			lastBlockHash := lastBlock.GetHeader().GetBlock_ID().GetBlockHash()
			lbh = lastBlockHash
			lbHeight := lastBlock.GetHeader().GetHeight()

			log.Info().Msgf("Last Block Hash: %v", lbh)
			log.Info().Msgf("Height: %v", lbHeight)

			s := lastBlock.GetHeader().GetTime().GetSeconds()
			ts := time.Unix(s, 0)
			log.Info().Msgf("Timestamp: %v", ts)

			var stateRootHex cmn.HexBytes
			stateRoot = lastBlock.GetHeader().GetStateRoot()
			stateRootHex = stateRoot
			log.Info().Msgf("State root: %v", stateRootHex)

		}
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		if *supervisorFlag {
			supsvc.SetEnv(env)
			supsvc.SetWaitTime(waitTime)
			supsvc.SetNoOfPeersInGroup(noOfPeersInGroup)
			supsvc.SetStateRoot(stateRoot)
			supsvc.SetBackup(backup)
			// Check for deactivated validators and remove them from supervisor list
			if supsvc.Validator != nil && len(supsvc.Validator) > 0 {
				for _, v := range supsvc.Validator {
					if !net.ConnectionStateExists(v.Address) {
						supsvc.RemoveValidator(v.Address)
					}
				}
			}

			lastBlock := blockchainSvc.GetLastBlock()
			stateRoot = lastBlock.GetHeader().GetStateRoot()
			supsvc.SetStateRoot(stateRoot)
			baseBlock, err := supsvc.ProcessTxs(lastBlock, net)
			if err != nil {
				log.Error().Msg(err.Error())
			}

			if baseBlock != nil {
				//log.Info().Msgf("Block Detail: %v", baseBlock)
				err = blockchainSvc.AddBaseBlock(baseBlock)
				if err != nil {
					log.Error().Msgf("Failed to Add Base Block: %v", err)
				}

				var bbh, pbbh cmn.HexBytes
				pbbh = baseBlock.Header.LastBlockID.BlockHash
				bbh = baseBlock.GetHeader().GetBlock_ID().GetBlockHash()
				log.Info().Msg("New Block Added")
				log.Info().Msgf("Block Id: %v", bbh.String())
				log.Info().Msgf("Last Block Id: %v", pbbh.String())
				log.Info().Msgf("Block Height: %v", baseBlock.GetHeader().GetHeight())

				s := lastBlock.GetHeader().GetTime().GetSeconds()
				ts := time.Unix(s, 0)
				log.Info().Msgf("Timestamp : %v", ts)

				var stateRoot cmn.HexBytes
				stateRoot = baseBlock.GetHeader().GetStateRoot()
				log.Info().Msgf("State root : %v", stateRoot)
			}
		} else {
			validatorProcessor(net, reader, peers)
		}

	}
}

// validatorProcessor checks and validates all the new child blocks
func validatorProcessor(net *network.Network, reader *bufio.Reader, peers []string) {
	ctx := network.WithSignMessage(context.Background(), true)
	if firstPingFromValidator == 0 {
		supervisorClient, err := net.Client(peers[0])
		if err != nil {
			log.Printf("unable to get supervisor client: %+v", err)
			return
		}
		reply, err := supervisorClient.Request(ctx, &blockProtobuf.ConnectionMessage{Message: "Connection established with Validator"})
		if err != nil {
			log.Printf("unable to request from client: %+v", err)
			return
		}
		fmt.Println("Supervisor reply: " + reply.String())
		firstPingFromValidator++
		return
	}

	// Check if a new child block has arrived
	if isChildBlockReceivedByValidator {
		vService := validator.Validator{}

		//Get all the transaction data included in the child block
		txsData := mcb.GetChildBlock().GetTxsData()
		if txsData == nil {
			fmt.Println("No txsData")
			isChildBlockReceivedByValidator = false
			return
		}
		txs := txsData.Tx

		//Get Root hash of the transactions
		cbRootHash := mcb.GetChildBlock().GetHeader().GetRootHash()
		err := vService.VerifyTxs(cbRootHash, txs)
		if err != nil {
			fmt.Println("Failed to verify transaction:", err)
			return
		}

		// Sign and vote the child block
		err = vService.Vote(net, net.Address, mcb)
		if err != nil {
			net.Broadcast(ctx, &blockProtobuf.ConnectionMessage{Message: "Failed to get vote"})
		}

		net.Broadcast(ctx, mcb)
		isChildBlockReceivedByValidator = false
	}
}
