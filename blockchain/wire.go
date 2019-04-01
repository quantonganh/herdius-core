package blockchain

import (
	"fmt"

	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/storage/db"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()
var badgerDB db.DB

func init() {

	cryptoAmino.RegisterAmino(cdc)
}

func LoadDB() {
	var dir string
	var dbName string
	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file not found...")
	} else {
		dir = viper.GetString("development.chaindbpath")
		dbName = viper.GetString("development.badgerDb")
		fmt.Println("Dir Name :" + dir)
		fmt.Println("dbName Name :" + dbName)
	}

	badgerDB = db.NewDB(dbName, db.GoBadgerBackend, dir)
}
