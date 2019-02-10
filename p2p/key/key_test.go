package key

import (
	"os"
	"path/filepath"
	"testing"

	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/stretchr/testify/assert"
)

func TestLoadOrGenNodeKey(t *testing.T) {
	filePath := filepath.Join(os.TempDir(), cmn.RandStr(12)+"_peer_id.json")

	nodeKey, err := LoadOrGenNodeKey(filePath)
	assert.Nil(t, err)

	nodeKey2, err := LoadOrGenNodeKey(filePath)
	assert.Nil(t, err)

	assert.Equal(t, nodeKey, nodeKey2)

	os.RemoveAll(filePath)
}
