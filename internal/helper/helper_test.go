package helper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomString(t *testing.T) {
	assert.Less(t, Numeric.Intn(10), 10)
	assert.Equal(t, len(AlphabetNumeric.Generate(16)), 16)
	Numeric.Uint32()
	Numeric.Uint64()
}

func TestUniq(t *testing.T) {
	assert.ElementsMatch(t, Uniq([]int{1, 3, 5, 7, 7, 9}), []int{1, 3, 5, 7, 9})
	assert.ElementsMatch(t, Uniq([]string{"ming", "ming", "shi"}), []string{"ming", "shi"})
}
