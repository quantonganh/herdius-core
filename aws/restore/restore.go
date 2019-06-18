package restore

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/herdius/herdius-core/config"
)

type RestorerI interface {
	Restore() error
	testCompleteChainRemote() (bool, error)
	clearOldChain() error
	downloadChain() error
	replayChain() error
}

type Restorer struct {
	statePath       string
	chainPath       string
	s3bucket        string
	heightToRestore int
	s3              *s3.S3
}

func NewRestorer(env string, height int) RestorerI {
	detail := config.GetConfiguration(env)
	s := s3.New(session.New())
	return Restorer{
		statePath:       detail.StateDBPath,
		chainPath:       detail.ChainDBPath,
		s3bucket:        detail.S3Bucket,
		heightToRestore: height,
		s3:              s,
	}
}

// Restore retrieves and procceses an entire blockchain stored in S3
// into the Supervisor's local blockchain and statedb
func (r Restorer) Restore() error {
	succ, err := r.testCompleteChainRemote()
	if err != nil {
		err = fmt.Errorf("restore failed while trying to test remote chain: %v", err)
	}
	if !succ {
		err = fmt.Errorf("could not restore chain from backup, specified chain in S3 is invalid")
	}

	log.Println("Clearing old chain")
	err = r.clearOldChain()
	if err != nil {
		return fmt.Errorf("restore failed while trying to clean old chain: %v", err)
	}

	err = r.downloadChain()
	if err != nil {
		return fmt.Errorf("restore failed while trying to download backed up chain: %v", err)
	}

	err = r.replayChain()
	if err != nil {
		return fmt.Errorf("restore failed while trying to replay chain: %v", err)
	}

	return fmt.Errorf("unable to restore chain entirely, but reached end of Remote()")
}

func (r Restorer) testCompleteChainRemote() (bool, error) {
	listParams := &s3.ListObjectsV2Input{
		Bucket:  aws.String(r.s3bucket),
		Prefix:  aws.String("0/blocks/"),
		MaxKeys: aws.Int64(1),
	}
	listResult, err := r.s3.ListObjectsV2(listParams)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve list of S3 objects: %v", err)
	}
	if len(listResult.Contents) != 1 {
		return false, fmt.Errorf("failed to find base block in S3 (block height = 0)")
	}
	log.Printf("root base block: %+v", *listResult.Contents[0].Key)
	key := *listResult.Contents[0].Key
	downloadParams := &s3.GetObjectInput{
		Bucket: aws.String(r.s3bucket),
		Key:    aws.String(key),
	}

	for i := 0; i < r.heightToRestore; i++ {
		downloadParams.Key = aws.String(key)
		downResult, err := r.s3.GetObject(downloadParams)
		if err != nil {
			return false, fmt.Errorf("failed to download S3 objects (height=%v, key=%v): %v", i, key, err)
		}
		key, err = getKeyFromDownload(i+1, downResult)
		if err != nil {
			return false, fmt.Errorf("failed to get key from prior block download (height=%v): %v", i, err)
		}

	}
	return true, nil
}

func (r Restorer) clearOldChain() error {
	return os.RemoveAll(r.chainPath)
}

func (r Restorer) downloadChain() error {
	return nil
}

func (r Restorer) replayChain() error {
	return nil
}

func getKeyFromDownload(i int, obj *s3.GetObjectOutput) (string, error) {
	log.Println("body content length:", *obj.ContentLength)
	body := make([]byte, *obj.ContentLength)
	_, err := obj.Body.Read(body)
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return "", fmt.Errorf("failed read body of s3 object output (i=%v): %v", i, err)
	}

	type b struct {
		Header struct {
			BlockID struct {
				BlockHash string `json:"blockHash"`
			} `json:"block_ID"`
		} `json:"header"`
	}
	var block b
	err = json.Unmarshal(body, &block)
	if err != nil {
		return "", fmt.Errorf("failed to parse json from request body (i=%v): %v", i, err)
	}
	if block.Header.BlockID.BlockHash == "" {
		return "", fmt.Errorf("request body json contains no blockhash (i=%v)", i)
	}
	log.Printf("block parsed: %+v", block)
	hash := block.Header.BlockID.BlockHash

	return fmt.Sprintf("%v/blocks/%v", i, hash), nil
}
