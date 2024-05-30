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

func TestFilter(t *testing.T) {
	var arr = []int{1, 2, 3, 4}
	arr = Filter(arr, func(i int, item int) bool {
		return item%2 == 0
	})
	assert.ElementsMatch(t, arr, []int{2, 4})
}

func TestClone(t *testing.T) {
	var arr = []int{1, 3, 5}
	assert.ElementsMatch(t, Clone(arr), arr)
}

func TestSplit(t *testing.T) {
	assert.ElementsMatch(t, []string{"api", "v1"}, Split("/api/v1", "/"))
	assert.ElementsMatch(t, []string{"api", "v1"}, Split("/api/v1/", "/"))
	assert.ElementsMatch(t, []string{"ming", "hong", "hu"}, Split("ming/ hong/ hu", "/"))
	assert.ElementsMatch(t, []string{"ming", "hong", "hu"}, Split("/ming/ hong/ hu/ ", "/"))
	assert.ElementsMatch(t, []string{"ming", "hong", "hu"}, Split("\nming/ hong/ hu\n", "/"))
}
