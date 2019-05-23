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
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/config"
)

// BackupBaseBlock takes a single block
func BackupBaseBlock(env string, lastblock, baseBlock *protobuf.BaseBlock) error {
	sess := session.Must(session.NewSession())

	bBlock, err := json.Marshal(baseBlock)
	if err != nil {
		return fmt.Errorf("cannot convert baseBlock to json: %v", err)
	}
	fileName := "tmpfile.txt"
	err = ioutil.WriteFile(fileName, bBlock, 0644)
	if err != nil {
		return fmt.Errorf("cannot write to %v: %v", fileName, err)
	}
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("cannot read from %v: %v", fileName, err)
	}

	bucket := config.GetConfiguration(env).S3Bucket
	heightStr := strconv.Itoa(int(baseBlock.Header.Height))
	timeStamp := strconv.Itoa(int(time.Now().Unix()))

	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(fmt.Sprintf("%v-%v", heightStr, timeStamp)),
		Body:                 file,
		ServerSideEncryption: aws.String("AES256"),
		Tagging:              aws.String(fmt.Sprintf("height=%v&timestamp=%v", heightStr, timeStamp)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	log.Println("Uploaded file to S3:", result)
	return nil
}
