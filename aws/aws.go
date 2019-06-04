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

// BackuperI ....
type BackuperI interface {
	TryBackupBaseBlock(*protobuf.BaseBlock, *protobuf.BaseBlock) (bool, error)
	BackupNeededBaseBlocks(*protobuf.BaseBlock) error
	backupToS3(*s3manager.Uploader, *protobuf.BaseBlock) (*s3manager.UploadOutput, error)
	findInS3(*s3.S3, *protobuf.BaseBlock) (bool, error)
}

// Backuper ...
type Backuper struct {
	Session *session.Session
	Bucket  string
}

// NewBackuper creates a standard AWS SDK session
func NewBackuper(env string) BackuperI {
	bucket := config.GetConfiguration(env).S3Bucket
	sess := session.New()
	b := &Backuper{
		Bucket:  bucket,
		Session: sess,
	}
	return b
}

// TryBackupBaseBlock takes a single block, returns true if able and successfully backup, false if business logic makes backup
// not useful, and errors if attempted backup fails
func (b *Backuper) TryBackupBaseBlock(lastBlock, baseBlock *protobuf.BaseBlock) (bool, error) {
	svc := s3.New(b.Session)
	found, err := b.findInS3(svc, lastBlock)
	if err != nil {
		return false, fmt.Errorf("failure searching S3 for previous block backup: %v", err)
	}
	if !found {
		return false, nil
	}

	uploader := s3manager.NewUploader(b.Session)
	res, err := b.backupToS3(uploader, baseBlock)
	if err != nil {
		return false, fmt.Errorf("could not backup new base block to S3: %v", err)
	}
	log.Println("Uploaded base block file to S3:", res.Location)
	return true, nil
}

// BackupNeededBaseBlocks iteratively goes through the entire blockchain and pushes up the contents of each block into S3
func (b *Backuper) BackupNeededBaseBlocks(newBlock *protobuf.BaseBlock) error {
	cdc := amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)

	viper.SetConfigName("config")   // Config file name without extension
	viper.AddConfigPath("./config") // Path to config file
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("config file not found: %v", err)
	}

	bDB := blockchain.GetBlockchainDb()

	svc := s3.New(b.Session)
	uploader := s3manager.NewUploader(b.Session)
	res, err := b.backupToS3(uploader, newBlock)
	if err != nil {
		return fmt.Errorf("aborting: while trying to backup all needed base blocks, could not backup new base block to S3: %v", err)
	}
	log.Println("Block backed up to S3:", res.Location)

	var blockHash common.HexBytes
	added := 0
	maxThreads := 1000
	sem := make(chan bool, maxThreads)

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

			sem <- true
			go func(blockHash common.HexBytes) {
				defer func() { <-sem }()
				found, err := b.findInS3(svc, block)
				if err != nil {
					log.Println("nonfatal: while attempting full chain backup, error while searching for block", err)
					return
				}
				if found {
					log.Printf("block found in s3 while backing up entire chain: %v-%v", block.Header.Height, blockHash)
					return
				}
				log.Printf("block not found in S3, backing up: %v-%v", block.Header.Height, blockHash)
				defer func() {
					log.Println("Blocks backed up to S3:", added)
				}()
				res, err := b.backupToS3(uploader, block)
				if err != nil {
					log.Println("nonfatal: could not backup base block to S3:", err)
					return
				}
				log.Println("Block backed up to S3:", res.Location)
				added++
			}(blockHash)
		}
		return nil
	})
	log.Println("adding back to cap on semaphore")
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	log.Println("finished adding to cap semaphore; returning")
	return err
}

// findInS3 searches for a given baseBlock in S3 in the given bucket
func (b *Backuper) findInS3(svc *s3.S3, baseBlock *protobuf.BaseBlock) (bool, error) {
	var blockHash common.HexBytes
	blockHashBz := baseBlock.GetHeader().GetBlock_ID().GetBlockHash()
	blockHash = blockHashBz
	blockHeight := strconv.FormatInt(baseBlock.Header.Height, 10)
	prefixPattern := fmt.Sprintf("%v-%v", blockHeight, blockHash)
	search := &s3.ListObjectsV2Input{
		Bucket: aws.String(b.Bucket),
		Prefix: aws.String(prefixPattern),
	}
	result, err := svc.ListObjectsV2(search)
	if err != nil {
		return false, fmt.Errorf("could not list previous block in S3: %v", err)
	}
	if len(result.Contents) <= 0 {
		return false, nil
	}
	return true, nil
}

// backupToS3 backs up a single baseBlock to S3 in the given bucket
func (b *Backuper) backupToS3(uploader *s3manager.Uploader, baseBlock *protobuf.BaseBlock) (result *s3manager.UploadOutput, err error) {
	fileName := "tmpfile.txt"

	// TODO CHANGE AWAY FROM MY OWN HOME DIR
	tmpFile, err := ioutil.TempFile("", fileName)
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
		Bucket:               aws.String(b.Bucket),
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
