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
func BackupBaseBlock(env string, lastBlock, baseBlock *protobuf.BaseBlock) error {
	sess := session.Must(session.NewSession())

	// TODO CHECK lastBlock vs. S3.Get(baseBlock.Height-1)
	// TODO CHECK lastBlock vs. S3.Get(baseBlock.Height-1)
	var blockHash common.HexBytes
	lastBlockHash := lastBlock.GetHeader().GetBlock_ID().GetBlockHash()
	blockHash = lastBlockHash
	lastBlockHeight := strconv.FormatInt(lastBlock.Header.Height, 10)
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
	if err != nil {
		return fmt.Errorf("could not list previous block in S3: %v", err)
	}
	log.Printf("result of search:\n%+v", result)
	// TODO CHECK lastBlock vs. S3.Get(baseBlock.Height-1)
	// TODO CHECK lastBlock vs. S3.Get(baseBlock.Height-1)

	uploader := s3manager.NewUploader(sess)
	res, err := backupToS3(uploader, bucket, baseBlock)
	if err != nil {
		return fmt.Errorf("could not backup new base block to S3: %v", err)
	}
	log.Println("Uploaded base block file to S3:", res.Location)
	return nil
}

// BackupAllBaseBlocks iteratively goes through the entire blockchain and pushes up the contents of each block into S3
func BackupAllBaseBlocks() (err error) {
	cdc := amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)

	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err = viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Config file not found: %v", err)
	}

	bDB := blockchain.GetBlockchainDb()

	err = bDB.GetBadgerDB().View(func(txn *badger.Txn) error {
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
			log.Printf("\nlastblock:\n%+v", lb)
		}
		return nil
	})
	return err
}

func backupToS3(uploader *s3manager.Uploader, bucket string, baseBlock *protobuf.BaseBlock) (result *s3manager.UploadOutput, err error) {
	fileName := "tmpfile.txt"

	// TODO CHANGE AWAY FROM MY OWN HOME DIR
	tmpFile, err := ioutil.TempFile("/Users/bittelc/go/src/github.com/herdius/herdius-blockchain-api", fileName)
	if err != nil {
		return nil, fmt.Errorf("cannot create tmpfile %v: %v", fileName, err)
	}
	defer os.Remove(tmpFile.Name())

	bBlock, err := json.Marshal(baseBlock)
	if err != nil {
		return nil, fmt.Errorf("cannot convert baseBlock to json: %v", err)
	}

	_, err = tmpFile.Write(bBlock)
	if err != nil {
		return nil, fmt.Errorf("cannot write to: tmpfile %v: %v", fileName, err)
	}
	defer func() {
		errF := tmpFile.Close()
		if err != nil && errF != nil {
			err = fmt.Errorf("%v: %v", err, errF)
		}
		if errF != nil {
			err = fmt.Errorf("cannot close tmpfile %v: %v", fileName, errF)
		}
	}()
	err = tmpFile.Sync()
	if err != nil {
		return nil, fmt.Errorf("cannot sync file %v: %v", fileName, err)
	}
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot seek file %v: %v", fileName, err)
	}

	heightStr := strconv.Itoa(int(baseBlock.Header.Height))
	timeStamp := strconv.Itoa(int(time.Now().Unix()))
	var blockHash common.HexBytes
	blockHash = baseBlock.GetHeader().GetBlock_ID().GetBlockHash()

	result, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(fmt.Sprintf("%v-%v", heightStr, blockHash)),
		Body:                 tmpFile,
		ServerSideEncryption: aws.String("AES256"),
		Tagging:              aws.String(fmt.Sprintf("height=%v&timestamp=%v&blockhash=%v", heightStr, timeStamp, blockHash)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %v", err)
	}
	return result, err
}
