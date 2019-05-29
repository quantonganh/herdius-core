package aws

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
		"github.com/stretchr/testify/assert"
			"github.com/herdius/herdius-core/aws/aws_mocks"
)

func TestTryBackupBaseBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mBackuper := aws_mocks.NewMockBackuperI(ctrl)
	mBackuper.
		EXPECT().
		TryBackupBaseBlock(nil, nil).
		Return(true, nil).
		AnyTimes()
	mBackuper.
		EXPECT().
		BackupNeededBaseBlocks(gomock.Eq(nil)).
		Return(nil).
		AnyTimes()
		//
	succ, err := mBackuper.TryBackupBaseBlock(nil, nil)
	assert.True(t, succ)
	assert.NoError(t, err)

}
