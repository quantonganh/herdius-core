package mempool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMemPool(t *testing.T) {
	m := GetMemPool()
	assert.Implements(t, (*Service)(nil), m)
}
