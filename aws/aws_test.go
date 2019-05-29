package aws

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/herdius/herdius-core/aws/aws_mocks"
)

func TestTryBackupBaseBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := aws_mocks.NewMockBackuperI(ctrl)
	m.
		EXPECT().
		TryBackupBaseBlock(gomock.Eq(nil), gomock.Eq(nil)).
		Return(true, nil).
		AnyTimes()

	TryBackupBaseBlock(m)

}
