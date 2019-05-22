package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/herdius/herdius-core/blockchain/protobuf"
)

// BackupBaseBlock takes a single block
func BackupBaseBlock(lastblock, baseBlock *protobuf.BaseBlock) error {
	log.Println("hello world")

	sess := session.Must(session.NewSession())
	sthree := s3.New(sess)

	heightStr := strconv.Itoa(int(baseBlock.Header.Height))
	timeStamp := strconv.Itoa(int(time.Now().Unix()))
	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String("baseblock_backups"),
		Key:                  aws.String(fmt.Sprintf("%v-%v", heightStr, timeStamp)),
		Body:                 baseBlock,
		ServerSideEncryption: aws.String("AES256"),
		Tagging:              aws.String(fmt.Sprintf("height=%v&timestamp=%v", heightStr, timeStamp)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	return nil
}
