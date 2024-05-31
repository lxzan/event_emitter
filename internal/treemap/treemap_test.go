package treemap

import (
	"fmt"
	"github.com/lxzan/event_emitter/internal/helper"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"
)

func TestTreeMap(t *testing.T) {
	var topics []string
	var count = 100000
	for i := 0; i < count; i++ {
		var topic = strings.Builder{}
		topic.Write(helper.A2F.Generate(3))
		topic.WriteString(".")
		topic.Write(helper.A2F.Generate(3))
		topic.WriteString(".")
		topic.Write(helper.A2F.Generate(3))
		topics = append(topics, topic.String())
	}
	topics = helper.Uniq(topics)
	count = len(topics)

	var tm = New[uint8](".")
	for _, item := range topics {
		tm.Put(item, 1)
	}
	for i := 0; i < 100; i++ {
		var index = helper.A2F.Intn(count)
		var s = topics[index]
		var list = helper.Split(s, ".")

		expected := helper.Filter(helper.Clone(topics), func(i int, v string) bool {
			temp := helper.Split(v, ".")
			return temp[1] == list[1] && temp[2] == list[2]
		})
		actual := make([]string, 0)
		tm.Match(fmt.Sprintf("*.%s.%s", list[1], list[2]), func(k string, v uint8) {
			actual = append(actual, k)
		})
		assert.ElementsMatch(t, expected, actual)

		expected = helper.Filter(helper.Clone(topics), func(i int, v string) bool {
			temp := helper.Split(v, ".")
			return temp[0] == list[0] && temp[2] == list[2]
		})
		actual = make([]string, 0)
		tm.Match(fmt.Sprintf("%s.*.%s", list[0], list[2]), func(k string, v uint8) {
			actual = append(actual, k)
		})
		assert.ElementsMatch(t, expected, actual)

		expected = helper.Filter(helper.Clone(topics), func(i int, v string) bool {
			temp := helper.Split(v, ".")
			return temp[0] == list[0] && temp[1] == list[1]
		})
		actual = make([]string, 0)
		tm.Match(fmt.Sprintf("%s.%s.*", list[0], list[1]), func(k string, v uint8) {
			actual = append(actual, k)
		})
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestTreeMap_Put(t *testing.T) {
	var tm = New[uint8](".")
	tm.Put("aha", 1)
	tm.Put(".", 1)
}

func TestTreeMap_Get(t *testing.T) {
	var tm = New[int](".")
	tm.Put("aha", 1)
	tm.Put("oh", 2)
	v, _ := tm.Get("aha")
	assert.Equal(t, v, 1)
	v, _ = tm.Get("oh")
	assert.Equal(t, v, 2)
	_, ok := tm.Get("~aha")
	assert.False(t, ok)
}

func TestTreeMap_Match(t *testing.T) {
	var tm = New[uint8](".")
	tm.Put("api.v1.actions.eat", 1)
	tm.Put("api.v1.actions", 2)
	var wg = new(sync.WaitGroup)
	wg.Add(2)
	var sum = uint8(0)
	tm.Match("api.v1.*.eat", func(k string, v uint8) {
		sum += v
		wg.Done()
	})
	tm.Match("api.v1.*", func(k string, v uint8) {
		sum += v
		wg.Done()
	})
	tm.Match("api.v1.any", func(k string, v uint8) {
		sum += v
		wg.Done()
	})
	wg.Wait()
	assert.Equal(t, sum, uint8(3))
}
