package types_test

import (
	"testing"

	"github.com/convox/praxis/types"

	"github.com/stretchr/testify/assert"
)

func TestId(t *testing.T) {
	id := types.Id("A", 10)

	assert.Equal(t, "A", id[0:1])
	assert.Len(t, id, 10)
}
