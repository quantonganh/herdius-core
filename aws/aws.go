package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	backupBlock(*s3manager.Uploader, *protobuf.BaseBlock) (*s3manager.UploadOutput, error)
	backupStateDB(*s3manager.Uploader, int64) error
	findBlockInS3(*s3.S3, *protobuf.BaseBlock) (bool, error)
}

// Backuper ...
type Backuper struct {
	Session      *session.Session
	Bucket       string
	StateDirPath string
	BackupPath   string
	timeStamp    string
}

// NewBackuper creates a standard AWS SDK session
func NewBackuper(env string) BackuperI {
	bucket := config.GetConfiguration(env).S3Bucket
	sdp := config.GetConfiguration(env).StateDBPath
	timeStamp := strconv.Itoa(int(time.Now().Unix()))
	sess := session.New()
	return &Backuper{
		Bucket:       bucket,
		Session:      sess,
		StateDirPath: sdp,
		timeStamp:    timestamp,
	}
}

// TryBackupBaseBlock takes a single block, returns true if able and successfully backup, false if business logic makes backup
// not useful, and errors if attempted backup fails
func (b *Backuper) TryBackupBaseBlock(lastBlock, baseBlock *protobuf.BaseBlock) (bool, error) {
	svc := s3.New(b.Session)
	found, err := b.findBlockInS3(svc, lastBlock)
	if err != nil {
		return false, fmt.Errorf("failure searching S3 for previous block backup: %v", err)
	}
	if !found {
		return false, nil
	}

	uploader := s3manager.NewUploader(b.Session)
	res, err := b.backupBlock(uploader, baseBlock)
	if err != nil {
		return false, fmt.Errorf("could not backup new base block to S3: %v", err)
	}
	log.Println("Uploaded base block file to S3:", res.Location)

	err = b.backupStateDB(uploader, baseBlock.Header.Height)
	if err != nil {
		return false, fmt.Errorf("could not backup State DB to S3: %v", err)
	}
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
	res, err := b.backupBlock(uploader, newBlock)
	if err != nil {
		return fmt.Errorf("while trying to backup all needed base blocks, could not backup new base blocks to S3: %v", err)
	}
	log.Println("Block backed up to S3:", res.Location)

	var blockHash common.HexBytes
	added, failed, maxThread := 0, 0, 200
	height := int64(0)
	sem := make(chan struct{}, maxThread)

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

			sem <- struct{}{}
			go func(blockHash common.HexBytes) {
				defer func() { <-sem }()
				found, err := b.findBlockInS3(svc, block)
				if err != nil {
					log.Println("Nonfatal: while attempting full chain backup, error while searching for block", err)
					failed++
					return
				}
				if found {
					log.Printf("Block found in s3 while backing up entire chain: %v", b.BackupPath)
					return
				}
				log.Printf("Block not found in S3, backing up: %v", b.BackupPath)
				res, err := b.backupBlock(uploader, block)
				if err != nil {
					log.Println("Nonfatal: could not backup base block to S3:", err)
					failed++
					return
				}
				if block.Header.Height > height {
					height = block.Header.Height
				}
				log.Println("Block backed up to S3:", res.Location)
				added++
			}(blockHash)
		}
		return nil
	})
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	log.Printf("Finished backing up all blocks; added blocks: %v, chain height: %v, blocks failed to backup: %v", added, height, failed)
	err = b.backupStateDB(uploader, height)
	if err != nil {
		return fmt.Errorf("Nonfatal: could not backup state DB to S3: %v", err)
	}
	return nil
}

// findInS3 searches for a given baseBlock in S3 in the given bucket
func (b *Backuper) findBlockInS3(svc *s3.S3, baseBlock *protobuf.BaseBlock) (bool, error) {
	var blockHash common.HexBytes
	blockHashBz := baseBlock.GetHeader().GetBlock_ID().GetBlockHash()
	blockHash = blockHashBz
	blockHeight := strconv.FormatInt(baseBlock.Header.Height, 10)
	b.BackupPath = fmt.Sprintf("%v/blocks/%v", blockHeight, blockHash)
	search := &s3.ListObjectsV2Input{
		Bucket: aws.String(b.Bucket),
		Prefix: aws.String(b.BackupPath),
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

// backupBloc backs up a single baseBlock to S3 in the given Backuper.Bucket
func (b *Backuper) backupBlock(uploader *s3manager.Uploader, baseBlock *protobuf.BaseBlock) (result *s3manager.UploadOutput, err error) {
	fileName := "tmpfile.txt"

	tmpFile, err := ioutil.TempFile("", fileName)
	if err != nil {
		return nil, fmt.Errorf("cannot create tmpfile: %v ", err)
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
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot seek file %v: %v", fileName, err)
	}

	heightStr := strconv.Itoa(int(baseBlock.Header.Height))
	var blockHash common.HexBytes
	blockHash = baseBlock.GetHeader().GetBlock_ID().GetBlockHash()
	b.BackupPath = fmt.Sprintf("%v/blocks/%v", heightStr, blockHash)

	result, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(b.Bucket),
		Key:                  aws.String(b.BackupPath),
		Body:                 tmpFile,
		ServerSideEncryption: aws.String("AES256"),
		Tagging:              aws.String(fmt.Sprintf("height=%v&timestamp=%v&blockhash=%v", heightStr, timeStamp, blockHash)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %v", err)
	}
	return result, err
}

func (b *Backuper) backupStateDB(uploader *s3manager.Uploader, height int64) error {
	w := walker{
		uploader:     uploader,
		uploadBucket: b.Bucket,
	}
	err := w.setUploadTags(height, blockHash)
	if err != nil {
		return fmt.Errorf("couldn't set upload tags: %v", err)
	}

	err = filepath.Walk(b.StateDirPath, w.walk)
	if err != nil {
		return fmt.Errorf("couldn't walk dir: %v", err)
	}
	log.Printf("State DB files uploaded: [%+v]", strings.Join(w.files, ", "))

	return nil
}

type walker struct {
	files         []string
	uploader      *s3manager.Uploader
	uploadPath    string
	uploadBucket  string
	uploadTagsStr string
}

func (w *walker) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("err walking (%q): %v", path, err)
	}
	if info.IsDir() {
		return nil
	}
	w.files = append(w.files, path)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("couldn't open file %v: %v", path, err)
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("couldn't get file info (%q): %v", path, err)
	}
	buffer := make([]byte, fileInfo.Size())
	_, err = f.Read(buffer)
	if err != nil {
		return fmt.Errorf("couldn't read from file (%q): %v", path, err)
	}
	err := w.setUploadPath(height, fileInfo.Name())
	if err != nil {
		return fmt.Errorf("couldn't set uploadPath (%q): %v", path, err)
	}

	_, err = w.uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(w.uploadBucket),
		Key:                  aws.String(w.uploadPath),
		Body:                 bytes.NewReader(buffer),
		ServerSideEncryption: aws.String("AES256"),
		Tagging:              aws.String(w.uploadTagsStr),
	})
	if err != nil {
		return fmt.Errorf("couldn't upload file (%q) to S3: %v", path, err)
	}
	return nil
}

func (w *walker) setUploadPath(height int64, fileName string) error {
	currentPath := fmt.Sprintf("herdius/statedb/CURRENT")
	cur, err := os.Open(currentPath)
	if err != nil {
		return fmt.Errorf("couldn't open CURRENT statedb file: %v", err)
	}
	defer cur.Close()
	contents, err := ioutil.ReadAll(cur)
	if err != nil {
		return fmt.Errorf("couldn't read contents from CURRENT statedb file: %v", err)
	}
	curStr := string(contents)

	w.uploadPath = fmt.Sprintf("%v/statedb/%v/%v/", height, curStr, fileName)
	return nil
}

func (w *walker) setUploadTags(height int64, timeStamp, blockHash string) {
	w.uploadTagsStr = fmt.Sprintf("height=%v&timestamp=%v&blockhash=%v", height, timeStamp, blockHash)
	return
}
