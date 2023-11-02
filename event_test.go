package event_emitter

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestEventEmitter_Publish(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var em = New(nil)
		var wg = &sync.WaitGroup{}
		wg.Add(2)

		suber1 := em.NewSubscriber()
		em.Subscribe(suber1, "test", func(msg any) {
			t.Logf("id=%d, msg=%v\n", suber1, msg)
			wg.Done()
		})
		em.Subscribe(suber1, "oh", func(msg any) {})

		suber2 := em.NewSubscriber()
		em.Subscribe(suber2, "test", func(msg any) {
			t.Logf("id=%d, msg=%v\n", suber2, msg)
			wg.Done()
		})

		err := em.Publish(context.Background(), "test", "hello!")
		assert.NoError(t, err)
		assert.Equal(t, em.CountSubscriberByTopic("test"), 2)
		assert.ElementsMatch(t, em.GetTopicsBySubId(1), []string{"test", "oh"})
		assert.ElementsMatch(t, em.GetTopicsBySubId(2), []string{"test"})
		assert.NoError(t, em.Publish(context.Background(), "", 1))
		wg.Wait()
	})

	t.Run("timeout", func(t *testing.T) {
		var em = New(&Config{Concurrency: 1})
		var suber1 = em.NewSubscriber()
		em.Subscribe(suber1, "topic1", func(msg any) {
			time.Sleep(100 * time.Millisecond)
		})
		em.Subscribe(suber1, "topic2", func(msg any) {
			time.Sleep(100 * time.Millisecond)
		})

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		assert.NoError(t, em.Publish(ctx, "topic1", 1))
		assert.Error(t, em.Publish(ctx, "topic1", 2))
		assert.NoError(t, em.Publish(context.Background(), "topic2", 3))
	})

	t.Run("batch1", func(t *testing.T) {
		var em = New(nil)
		var count = 1000
		var wg = &sync.WaitGroup{}
		wg.Add(count)
		for i := 0; i < count; i++ {
			id := em.NewSubscriber()
			em.Subscribe(id, "greet", func(msg any) {
				wg.Done()
			})
		}
		var err = em.Publish(context.Background(), "greet", 1)
		assert.NoError(t, err)
		wg.Wait()
	})

	t.Run("batch2", func(t *testing.T) {
		var em = New(nil)
		var count = 1000
		var wg = &sync.WaitGroup{}
		wg.Add(count * (count + 1) / 2)

		var mapping = make(map[string]int)
		var mu = &sync.Mutex{}
		for i := 0; i < count; i++ {
			topic := fmt.Sprintf("topic%d", i)
			for j := 0; j < i+1; j++ {
				em.Subscribe(int64(j), topic, func(msg any) {
					wg.Done()
					mu.Lock()
					mapping[topic]++
					mu.Unlock()
				})
			}
		}

		for i := 0; i < count; i++ {
			topic := fmt.Sprintf("topic%d", i)
			var err = em.Publish(context.Background(), topic, i)
			assert.NoError(t, err)
		}

		wg.Wait()
		for i := 0; i < count; i++ {
			topic := fmt.Sprintf("topic%d", i)
			assert.Equal(t, mapping[topic], count-i)
		}
	})
}

func TestEventEmitter_UnSubscribe(t *testing.T) {
	var em = New(&Config{Concurrency: 1})
	var suber1 = em.NewSubscriber()
	em.Subscribe(suber1, "topic1", func(msg any) {
		time.Sleep(100 * time.Millisecond)
	})
	em.Subscribe(suber1, "topic2", func(msg any) {
		time.Sleep(100 * time.Millisecond)
	})
	em.Subscribe(suber1, "topic3", func(msg any) {
		time.Sleep(100 * time.Millisecond)
	})
	assert.ElementsMatch(t, em.GetTopicsBySubId(suber1), []string{"topic1", "topic2", "topic3"})
	em.UnSubscribe(suber1, "topic1")
	assert.ElementsMatch(t, em.GetTopicsBySubId(suber1), []string{"topic2", "topic3"})
	em.UnSubscribeAll(suber1)
	assert.Zero(t, len(em.GetTopicsBySubId(suber1)))

	em.UnSubscribeAll(suber1)
	assert.Zero(t, em.CountSubscriberByTopic("topic0"))
}
