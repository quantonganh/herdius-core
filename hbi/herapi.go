package main

import (
	"context"
	"os/user"
	"strconv"
	"time"

	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/p2p/crypto"
	"github.com/herdius/herdius-core/p2p/types/opcode"

	keystore "github.com/herdius/herdius-core/p2p/key"

	"github.com/herdius/herdius-core/hbi/message"
	protoplugin "github.com/herdius/herdius-core/hbi/protobuf"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
	"github.com/herdius/herdius-core/p2p/network/discovery"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	port := 5555
	host := "localhost"

	peer := "tcp://localhost:3000"
	peers := make([]string, 1)
	peers = append(peers, peer)

	nodeAddress := host + ":" + strconv.Itoa(port)

	nodekey, err := keystore.LoadOrGenNodeKey(user.HomeDir + "/" + nodeAddress + "_peer_id.json")

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

	opcode.RegisterMessageType(opcode.Opcode(1112), &protobuf.ConnectionMessage{})
	opcode.RegisterMessageType(opcode.Opcode(1113), &protoplugin.BlockHeightRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1114), &protoplugin.BlockResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1115), &protoplugin.AccountRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1116), &protoplugin.AccountResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1117), &protoplugin.TransactionRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1118), &protoplugin.TransactionResponse{})

	builder := network.NewBuilder()
	builder.SetKeys(keys)
	builder.SetAddress(network.FormatAddress("tcp", host, uint16(port)))

	// // Register peer discovery plugin.
	builder.AddPlugin(new(discovery.Plugin))
	builder.AddPlugin(new(message.BlockMessagePlugin))

	net, err := builder.Build()
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	go net.Listen()
	defer net.Close()

	if len(peers) > 0 {
		net.Bootstrap(peers...)
	}

	ctx := network.WithSignMessage(context.Background(), true)

	net.Broadcast(ctx, &protoplugin.BlockHeightRequest{BlockHeight: 2})
	net.Broadcast(ctx, &protoplugin.AccountRequest{Address: "2orwhatever"})

	time.Sleep(2 * time.Second)

}
