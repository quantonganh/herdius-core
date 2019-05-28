package service

import (
	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/p2p/log"
	external "github.com/herdius/herdius-core/storage/exbalance"
	"github.com/herdius/herdius-core/storage/state/statedb"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()
var trie statedb.Trie
var accountStorage external.BalanceStorage

func init() {

	cryptoAmino.RegisterAmino(cdc)
}

//LoadStateDB loads the state trie db
func LoadStateDB(exBal external.BalanceStorage) {
	var dir string
	accountStorage = exBal
	viper.SetConfigName("config")       // Config file name without extension
	viper.AddConfigPath("../../config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Msgf("Config file not found: %v\n", err)
	} else {
		dir = viper.GetString("dev.statedbpath")
	}

	trie = statedb.GetState(dir)

}
