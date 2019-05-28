package aws

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dgraph-io/badger"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"

	"github.com/herdius/herdius-core/blockchain"
	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/config"
	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/libs/common"
)

// BackupBaseBlock takes a single block
func BackupBaseBlock(env string, lastBlock, baseBlock *protobuf.BaseBlock) (err error) {
	sess := session.Must(session.NewSession())

	// TODO CHECK lastBlock vs. S3.Get(baseBlock.Height-1)
	// TODO CHECK lastBlock vs. S3.Get(baseBlock.Height-1)
	var blockHash common.HexBytes
	lastBlockHash := lastBlock.GetHeader().GetBlock_ID().GetBlockHash()
	blockHash = lastBlockHash
	lastBlockHeight := string(lastBlock.Header.Height)
	prefixPattern := fmt.Sprintf("%v-%v", lastBlockHeight, blockHash)
	log.Println("prefixPattern:", prefixPattern)

	bucket := config.GetConfiguration(env).S3Bucket
	svc := s3.New(session.New())
	search := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(2),
		Prefix:  aws.String(prefixPattern),
	}
	result, err := svc.ListObjectsV2(search)
	log.Printf("result of search:\n%+v", result)
	// TODO CHECK lastBlock vs. S3.Get(baseBlock.Height-1)
	// TODO CHECK lastBlock vs. S3.Get(baseBlock.Height-1)

	bBlock, err := json.Marshal(baseBlock)
	if err != nil {
		return fmt.Errorf("cannot convert baseBlock to json: %v", err)
	}
	fileName := "tmpfile.txt"

	tmpFile, err := ioutil.TempFile("/Users/bittelc/go/src/github.com/herdius/herdius-blockchain-api", fileName)
	if err != nil {
		return fmt.Errorf("cannot create tmpfile %v: %v", fileName, err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(bBlock)
	if err != nil {
		return fmt.Errorf("cannot write to: tmpfile %v: %v", fileName, err)
	}
	defer func() {
		err = tmpFile.Close()
		if err != nil {
			err = fmt.Errorf("cannot close tmpfile %v: %v", fileName, err)
		}
	}()
	err = tmpFile.Sync()
	if err != nil {
		return fmt.Errorf("cannot sync file %v: %v", fileName, err)
	}
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("cannot seek file %v: %v", fileName, err)
	}

	heightStr := strconv.Itoa(int(baseBlock.Header.Height))
	timeStamp := strconv.Itoa(int(time.Now().Unix()))
	thisBlockHash := baseBlock.GetHeader().GetBlock_ID().GetBlockHash()
	blockHash = thisBlockHash

	uploader := s3manager.NewUploader(sess)
	result1, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(fmt.Sprintf("%v-%v", heightStr, blockHash)),
		Body:                 tmpFile,
		ServerSideEncryption: aws.String("AES256"),
		Tagging:              aws.String(fmt.Sprintf("height=%v&timestamp=%v&blockhash=%v", heightStr, timeStamp, blockHash)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	log.Println("Uploaded file to S3:", result1)
	return
}

// BackupAllBaseBlocks iteratively goes through the entire blockchain and pushes up the contents of each block into S3
func BackupAllBaseBlocks() (err error) {
	cdc := amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)

	//var badgerDB db.DB
	//var dir string
	//var dbName string
	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err = viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Config file not found: %v", err)
		//} else {
		//dir = viper.GetString("dev.chaindbpath")
		//dbName = viper.GetString("dev.badgerDb")
	}

	//badgerDB := db.NewDB(dbName, db.GoBadgerBackend, dir)
	bDB := blockchain.GetBlockchainDb()

	//err = bDB.GetBadgerDB().View(func(txn *badger.Txn) error {
	err = bDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return fmt.Errorf("cannot retrieve item value: %v", err)
			}
			lb := &protobuf.BaseBlock{}
			err = cdc.UnmarshalJSON(v, lb)
			if err != nil {
				return fmt.Errorf("cannot unmarshal db block into struct block: %v", err)
			}
			log.Println("\nlastblock:\n", lb)
		}
		return nil
	})
	return err
}
