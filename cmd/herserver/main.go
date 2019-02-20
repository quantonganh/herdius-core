package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"

	"os"
	"strconv"
	"strings"
	"time"

	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/blockchain/protobuf"
	cryptokey "github.com/herdius/herdius-core/crypto"
	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/herdius/herdius-core/p2p/crypto"
	keystore "github.com/herdius/herdius-core/p2p/key"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
	"github.com/herdius/herdius-core/p2p/network/discovery"
	"github.com/herdius/herdius-core/p2p/types/opcode"
	sup "github.com/herdius/herdius-core/supervisor/service"
	"github.com/herdius/herdius-core/supervisor/transaction"
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
var mcb = &protobuf.ChildBlockMessage{}

// firstPingFromValidator checks whether a connection is established betweer supervisor and validator.
// And it is used to send a message on established connection.
var firstPingFromValidator = 0

var nodeKeydir = "./supervisor/testdata/"

var t1 time.Time
var t2 time.Time

func init() {

	supsvc = &sup.Supervisor{}
	supsvc.SetWriteMutex()
	supsvc.ValidatorChildblock = make(map[string]*protobuf.BlockID, 0)
	supsvc.ChildBlock = make([]*protobuf.ChildBlock, 0)
	supsvc.VoteInfoData = make(map[string][]*protobuf.VoteInfo, 0)

	RegisterAminoService(cdc)
}

//RegisterAminoService registers Amino service for message (en/de) coding
func RegisterAminoService(cdc *amino.Codec) {
	cryptoAmino.RegisterAmino(cdc)

}

// HerdiusMessagePlugin will receive all trasnmitted messages.
type HerdiusMessagePlugin struct{ *network.Plugin }

var addresses = make([]string, 0)

// Receive handles each received message for both Supervisor and Validator
func (state *HerdiusMessagePlugin) Receive(ctx *network.PluginContext) error {

	switch msg := ctx.Message().(type) {
	case *protobuf.ConnectionMessage:
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
		supsvc.ValidatorChildblock[address] = &protobuf.BlockID{}
		mx.Unlock()
	case *protobuf.ChildBlockMessage:

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

					var stateRoot cmn.HexBytes
					stateRoot = baseBlock.GetHeader().GetStateRoot()
					log.Info().Msgf("State root : %v", stateRoot)
					// Once new base block is added to be block chain
					// do the following

					supsvc.ValidatorChildblock = make(map[string]*protobuf.BlockID, 0)
					supsvc.ChildBlock = make([]*protobuf.ChildBlock, 0)
					supsvc.VoteInfoData = make(map[string][]*protobuf.VoteInfo, 0)
					mcb = &protobuf.ChildBlockMessage{}
					voteCount = 0

					supsvc.StateRoot = []byte{0}
					t2 = time.Now()

					diff := t2.Sub(t1)

					log.Info().Msgf("Total time : %v", diff)

				}

			} else {
				log.Info().Msgf("<%s> Validator verification or signature verification failed: %v", ctx.Client().ID.Address, isVerified)
			}

		} else {
			isChildBlockReceivedByValidator = true
			noOfTxs := mcb.GetChildBlock().GetHeader().GetNumTxs()
			log.Info().Msgf("<%s> #Txs: %v", ctx.Client().ID.Address, noOfTxs)
		}

	}
	return nil
}

func main() {
	// process other flags
	portFlag := flag.Int("port", 3000, "port to listen to")
	hostFlag := flag.String("host", "localhost", "host to listen to")
	protocolFlag := flag.String("protocol", "tcp", "protocol to use (kcp/tcp)")
	peersFlag := flag.String("peers", "", "peers to connect to")
	supervisorFlag := flag.Bool("supervisor", false, "Is supervisor host?")
	groupSizeFlag := flag.Int("groupsize", 3, "# of peers in a validator group")
	flag.Parse()

	port := uint16(*portFlag)
	host := *hostFlag
	protocol := *protocolFlag
	peers := strings.Split(*peersFlag, ",")
	fmt.Printf("Peers info: %v\n", peers)
	noOfPeersInGroup := *groupSizeFlag

	// Generate or Load Keys

	nodeAddress := host + ":" + strconv.Itoa(*portFlag)

	nodekey, err := keystore.LoadOrGenNodeKey(nodeKeydir + nodeAddress + "_peer_id.json")

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

	opcode.RegisterMessageType(opcode.Opcode(1111), &protobuf.ChildBlockMessage{})
	opcode.RegisterMessageType(opcode.Opcode(1112), &protobuf.ConnectionMessage{})

	builder := network.NewBuilder()
	builder.SetKeys(keys)

	builder.SetAddress(network.FormatAddress(protocol, host, port))

	// Register peer discovery plugin.
	builder.AddPlugin(new(discovery.Plugin))

	// Add custom chat plugin.
	builder.AddPlugin(new(HerdiusMessagePlugin))

	net, err := builder.Build()
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	go net.Listen()

	if len(peers) > 0 {
		net.Bootstrap(peers...)
	}

	// As of now Databases will only be loaded for Supervisor.
	// Chain data and state information will be stored at supervisor's node.

	var stateRoot []byte
	if *supervisorFlag {

		blockchain.LoadDB()
		sup.LoadStateDB()
		blockchainSvc := &blockchain.Service{}

		lastBlock := blockchainSvc.GetLastBlock()

		if err != nil {
			log.Error().Msgf("Failed while getting last block: %v\n", err)
		} else {
			var lbh cmn.HexBytes
			lastBlockHash := lastBlock.GetHeader().GetBlock_ID().GetBlockHash()
			lbh = lastBlockHash
			lbHeight := lastBlock.GetHeader().GetHeight()

			log.Info().Msgf("Last Block Hash : %v", lbh)
			log.Info().Msgf("Height : %v", lbHeight)

			s := lastBlock.GetHeader().GetTime().GetSeconds()
			ts := time.Unix(s, 0)
			log.Info().Msgf("Timestamp : %v", ts)

			var stateRootHex cmn.HexBytes
			stateRoot = lastBlock.GetHeader().GetStateRoot()
			stateRootHex = stateRoot
			log.Info().Msgf("State root : %v", stateRootHex)

		}
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		if *supervisorFlag {

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
			supervisorProcessor(net, reader, stateRoot, noOfPeersInGroup)
		} else {
			validatorProcessor(net, reader)

		}

	}
}

//supervisorProcessor processes all the incoming transactions
func supervisorProcessor(net *network.Network, reader *bufio.Reader, stateRoot []byte, noOfPeersInGroup int) {
	log.Info().Msg("Please press 'y' to load transactions from file. ")
	input, _ := reader.ReadString('\n')

	// skip blank lines
	// Right now this process only handles the txs loaded from a file
	// TODO: It has to be implemented in a more generic way so that it can handle txs arriving in various ways
	if len(strings.TrimSpace(input)) == 0 || strings.TrimRight(input, "\n") != "y" {
		return
	}
	totalNoOfPeers := len(addresses)

	totalTXsToBeValidated := 3000
	numberOfTXsInEachBatch := 500
	numberOfBatches := totalTXsToBeValidated / numberOfTXsInEachBatch

	err := supsvc.CreateTxBatchesFromFile("./supervisor/testdata/txs.json", numberOfBatches, numberOfTXsInEachBatch, stateRoot)

	if err != nil {
		log.Error().Msgf("Failed while batching the transactions: %v", err)
		return
	}
	batches := *supsvc.TxBatches

	counter := 0
	var txService transaction.Service

	numOfValidatorGroups := calcNoOfGroups(totalNoOfPeers, noOfPeersInGroup)

	batchCount := 0
	previousBlockHash := make([]byte, 0)

	var broadcastCount int
	broadcastCount = 0
	t1 = time.Now()
	for _, batch := range batches {

		txService = transaction.TxService()
		for i := 0; i < numberOfTXsInEachBatch; i++ {
			txbz := batch[i]

			tx := transaction.Tx{}
			cdc.UnmarshalJSON(txbz, &tx)

			txService.AddTx(tx)
			counter++
		}

		txList := *(txService.GetTxList())

		cb := supsvc.CreateChildBlock(net, &txList, int64(batchCount), previousBlockHash)

		previousBlockHash = cb.GetHeader().GetBlockID().BlockHash

		supsvc.ChildBlock = append(supsvc.ChildBlock, cb)
		cbmsg := &protobuf.ChildBlockMessage{
			ChildBlock: cb,
		}

		if batchCount > 0 {
			cb.GetHeader().GetLastBlockID().BlockHash = previousBlockHash
		}

		ctx := network.WithSignMessage(context.Background(), true)

		if totalNoOfPeers <= noOfPeersInGroup {

			net.BroadcastByAddresses(ctx, cbmsg, addresses...)
			batchCount++
		} else {
			for i := broadcastCount; i < totalNoOfPeers; i++ {
				if i+noOfPeersInGroup <= totalNoOfPeers {

					net.BroadcastByAddresses(ctx, cbmsg, addresses[i:i+noOfPeersInGroup]...)
					i = (i + noOfPeersInGroup) - 1

				} else {
					net.BroadcastByAddresses(ctx, cbmsg, addresses[i:totalNoOfPeers]...)
					i = (i + noOfPeersInGroup) - 1
				}
				broadcastCount = i + 1
				break
			}
			batchCount++
		}

		//break for batch loop in case it equals number of validator groups

		if numOfValidatorGroups == batchCount {
			break
		}
	}

}

// validatorProcessor checks and validates all the new child blocks
func validatorProcessor(net *network.Network, reader *bufio.Reader) {

	ctx := network.WithSignMessage(context.Background(), true)
	if firstPingFromValidator == 0 {
		net.Broadcast(ctx, &protobuf.ConnectionMessage{Message: "Connection established"})
		firstPingFromValidator++
		return
	}

	// Check if a new child block has arrived
	if isChildBlockReceivedByValidator {
		vService := validator.Validator{}

		//Get all the transaction data included in the child block
		txs := mcb.GetChildBlock().GetTxsData().Tx

		//Get Root hash of the transactions
		cbRootHash := mcb.GetChildBlock().GetHeader().GetRootHash()
		err := vService.VerifyTxs(cbRootHash, txs)
		if err != nil {
			net.Broadcast(ctx, &protobuf.ConnectionMessage{Message: "Failed to verify the transactions"})
		}

		// Sign and vote the child block
		err = vService.Vote(net, net.Address, mcb)
		if err != nil {
			net.Broadcast(ctx, &protobuf.ConnectionMessage{Message: "Failed to get vote"})
		}

		net.Broadcast(ctx, mcb)
		isChildBlockReceivedByValidator = false
	}

}
func calcNoOfGroups(totalPeers, gpc int) int {
	// gpc : Group peer count
	if (totalPeers % gpc) != 0 {
		noOfGrps := totalPeers % gpc
		if noOfGrps < gpc {
			return (totalPeers / gpc) + 1
		}
		return noOfGrps + (totalPeers / gpc)

	}
	return (totalPeers / gpc)
}
