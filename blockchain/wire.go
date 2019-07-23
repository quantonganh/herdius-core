package blockchain

import (
	"fmt"

	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/storage/db"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"
)

var (
	cdc               = amino.NewCodec()
	badgerDB          db.DB
	blockHeightHashDB db.DB
)

func init() {

	cryptoAmino.RegisterAmino(cdc)
}

// LoadDB loads databases used by blockchain
func LoadDB() {
	var (
		dir, dbName           string
		blockDir, blockDBName string
	)
	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file not found...")
	} else {
		dir = viper.GetString("dev.chaindbpath")
		dbName = viper.GetString("dev.badgerDb")
		blockDir = viper.GetString("dev.blockdbpath")
		blockDBName = viper.GetString("dev.badgerDb")
	}

	badgerDB = db.NewDB(dbName, db.GoBadgerBackend, dir)
	blockHeightHashDB = db.NewDB(blockDBName, db.GoBadgerBackend, blockDir)
}
