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

// TryBackupBaseBlock takes a single block
func TryBackupBaseBlock(env string, lastBlock, baseBlock *protobuf.BaseBlock) (bool, error) {
	sess := session.Must(session.NewSession())
	svc := s3.New(session.New())
	bucket := config.GetConfiguration(env).S3Bucket
	found, err := findInS3(svc, bucket, lastBlock)
	if err != nil {
		return false, fmt.Errorf("failure searching S3 for previous block backup: %v", err)
	}
	if !found {
		return false, nil
	}

	uploader := s3manager.NewUploader(sess)
	res, err := backupToS3(uploader, bucket, baseBlock)
	if err != nil {
		return false, fmt.Errorf("could not backup new base block to S3: %v", err)
	}
	log.Println("Uploaded base block file to S3:", res.Location)
	return true, nil
}

// BackupNeededBaseBlocks iteratively goes through the entire blockchain and pushes up the contents of each block into S3
func BackupNeededBaseBlocks(env string) (int, error) {
	cdc := amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)

	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		return 0, fmt.Errorf("Config file not found: %v", err)
	}

	bDB := blockchain.GetBlockchainDb()

	sess := session.Must(session.NewSession())
	svc := s3.New(session.New())
	bucket := config.GetConfiguration(env).S3Bucket
	uploader := s3manager.NewUploader(sess)

	var blockHash common.HexBytes
	notFound := make(chan *protobuf.BaseBlock, 30)
	added := 0

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
			block := &protobuf.BaseBlock{}
			err = cdc.UnmarshalJSON(v, block)
			if err != nil {
				return fmt.Errorf("cannot unmarshal db block into struct block: %v", err)
			}
			blockHash = block.Header.Block_ID.BlockHash
			log.Printf("lastblock height-hash: %v-%v", block.Header.Height, blockHash)
			go func() {
				found, err := findInS3(svc, bucket, block)
				if err != nil {
					log.Println("nonfatal: while attempting full chain backup, unable to find block", err)
					return
				}
				if !found {
					log.Printf("not found in S3: %v-%v", block.Header.Height, blockHash)
					notFound <- block
				}
			}()
			go func() {
				for {
					select {
					case unbacked := <-notFound:
						log.Printf("not found in S3, beginning backup: %v-%v", unbacked.Header.Height, blockHash)
						res, err := backupToS3(uploader, bucket, unbacked)
						if err != nil {
							log.Printf("nonfatal: could not backup base block to S3: %v", err)
						}
						log.Println("Block backed up to S3:", res.Location)
						added++
					}
				}
			}()
		}
		close(notFound)
		return nil
	})
	return added, err
}

// findInS3 searches for a given baseBlock in S3 in the given bucket
func findInS3(svc *s3.S3, bucket string, baseBlock *protobuf.BaseBlock) (bool, error) {
	var blockHash common.HexBytes
	blockHashBz := baseBlock.GetHeader().GetBlock_ID().GetBlockHash()
	blockHash = blockHashBz
	blockHeight := strconv.FormatInt(baseBlock.Header.Height, 10)
	prefixPattern := fmt.Sprintf("%v-%v", blockHeight, blockHash)
	search := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(2),
		Prefix:  aws.String(prefixPattern),
	}
	result, err := svc.ListObjectsV2(search)
	if err != nil {
		return false, fmt.Errorf("could not list previous block in S3: %v", err)
	}
	if len(result.Contents) <= 0 {
		return false, fmt.Errorf("previous base block could not be found in S3")
	}
	return true, nil
}

// backupToS3 backs up a single baseBlock to S3 in the given bucket
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
