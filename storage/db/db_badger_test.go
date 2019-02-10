package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	cmn "github.com/herdius/herdius-core/libs/common"
	"github.com/stretchr/testify/require"
)

func TestNewBadgerDB(t *testing.T) {

	dirname, err := ioutil.TempDir(os.TempDir(), "badgerdb_test_")
	db, err := NewBadgerDB(dirname, dirname)
	require.Nil(t, err)
	defer os.RemoveAll(dirname)
	db.Close() // Close the db to release the lock
}

func BenchmarkRandomReadsWrites(b *testing.B) {
	dirname, err := ioutil.TempDir(os.TempDir(), "badgerdb_test_")
	b.StopTimer()

	numItems := int64(1000000)
	internal := map[int64]int64{}
	for i := 0; i < int(numItems); i++ {
		internal[int64(i)] = int64(0)
	}

	db, err := NewBadgerDB(dirname, dirname)
	require.Nil(b, err)
	defer os.RemoveAll(dirname)

	fmt.Println("ok, starting")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// Random Writes
		{
			idx := (int64(cmn.RandInt()) % numItems)
			internal[idx]++
			val := internal[idx]
			idxBytes := int642Bytes(int64(idx))
			valBytes := int642Bytes(int64(val))

			db.Set(
				idxBytes,
				valBytes,
			)
		}

		// Random Reads
		{
			idx := (int64(cmn.RandInt()) % numItems)
			val := internal[idx]
			idxBytes := int642Bytes(int64(idx))
			valBytes := db.Get(idxBytes)

			if val == 0 {
				if !bytes.Equal(valBytes, nil) {
					b.Errorf("Expected %v for %v, got %X",
						nil, idx, valBytes)
					break
				}
			} else {
				if len(valBytes) != 8 {
					b.Errorf("Expected length 8 for %v, got %X",
						idx, valBytes)
					break
				}
				valGot := bytes2Int64(valBytes)
				if val != valGot {
					b.Errorf("Expected %v for %v, got %v",
						val, idx, valGot)
					break
				}
			}
		}
	}
}

func int642Bytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytes2Int64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
