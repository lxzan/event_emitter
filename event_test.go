package event_emitter

import (
	"fmt"
	"github.com/lxzan/event_emitter/internal/helper"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestEventEmitter_Publish(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var em = New[Int64Subscriber](nil)
		var wg = &sync.WaitGroup{}
		wg.Add(2)

		suber1 := em.NewSubscriber()
		em.Subscribe(suber1, "test", func(subscriber Int64Subscriber, msg any) {
			t.Logf("id=%d, msg=%v\n", suber1, msg)
			wg.Done()
		})
		em.Subscribe(suber1, "oh", func(subscriber Int64Subscriber, msg any) {})

		suber2 := em.NewSubscriber()
		em.Subscribe(suber2, "test", func(subscriber Int64Subscriber, msg any) {
			t.Logf("id=%d, msg=%v\n", suber2, msg)
			wg.Done()
		})

		em.Publish("test", "hello!")
		assert.Equal(t, em.CountSubscriberByTopic("test"), 2)
		assert.ElementsMatch(t, em.GetTopicsBySubscriber(1), []string{"test", "oh"})
		assert.ElementsMatch(t, em.GetTopicsBySubscriber(2), []string{"test"})
		em.Publish("", 1)
		wg.Wait()
	})

	t.Run("timeout", func(t *testing.T) {
		var em = New[Int64Subscriber](&Config{})
		var suber1 = em.NewSubscriber()
		em.Subscribe(suber1, "topic1", func(subscriber Int64Subscriber, msg any) {
			time.Sleep(100 * time.Millisecond)
		})
		em.Subscribe(suber1, "topic2", func(subscriber Int64Subscriber, msg any) {
			time.Sleep(100 * time.Millisecond)
		})

		em.Publish("topic1", 1)
		em.Publish("topic1", 2)
		em.Publish("topic2", 3)
	})

	t.Run("batch1", func(t *testing.T) {
		var em = New[Int64Subscriber](nil)
		var count = 1000
		var wg = &sync.WaitGroup{}
		wg.Add(count)
		for i := 0; i < count; i++ {
			id := em.NewSubscriber()
			em.Subscribe(id, "greet", func(subscriber Int64Subscriber, msg any) {
				wg.Done()
			})
		}
		em.Publish("greet", 1)
		wg.Wait()
	})

	t.Run("batch2", func(t *testing.T) {
		var em = New[Int64Subscriber](nil)
		var count = 1000
		var wg = &sync.WaitGroup{}
		wg.Add(count * (count + 1) / 2)

		var mapping = make(map[string]int)
		var mu = &sync.Mutex{}
		for i := 0; i < count; i++ {
			topic := fmt.Sprintf("topic%d", i)
			for j := 0; j < i+1; j++ {
				em.Subscribe(Int64Subscriber(j), topic, func(subscriber Int64Subscriber, msg any) {
					wg.Done()
					mu.Lock()
					mapping[topic]++
					mu.Unlock()
				})
			}
		}

		for i := 0; i < count; i++ {
			topic := fmt.Sprintf("topic%d", i)
			em.Publish(topic, i)
		}

		wg.Wait()
		for i := 0; i < count; i++ {
			topic := fmt.Sprintf("topic%d", i)
			assert.Equal(t, mapping[topic], i+1)
		}
	})

	t.Run("batch3", func(t *testing.T) {
		var em = New[Int64Subscriber](&Config{BucketNum: 1})
		var count = 1000
		var mapping1 = make(map[string]int)
		var mapping2 = make(map[string]int)
		var mu = &sync.Mutex{}
		var subjects = make(map[string]uint8)
		var wg = &sync.WaitGroup{}

		for i := 0; i < count; i++ {
			var topics []string
			for j := 0; j < 100; j++ {
				topic := fmt.Sprintf("topic-%d", helper.Numeric.Intn(count))
				topics = append(topics, topic)
			}

			topics = helper.Uniq(topics)
			wg.Add(len(topics))
			for j, _ := range topics {
				var topic = topics[j]
				mapping1[topic]++
				subjects[topic] = 1
				em.Subscribe(Int64Subscriber(i), topic, func(subscriber Int64Subscriber, msg any) {
					mu.Lock()
					mapping2[topic]++
					mu.Unlock()
					wg.Done()
				})
			}
		}

		for k, _ := range subjects {
			em.Publish(k, "hello")
		}

		wg.Wait()
		for k, _ := range subjects {
			assert.Equal(t, mapping1[k], mapping2[k])
		}
	})
}

func TestEventEmitter_UnSubscribe(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var em = New[Int64Subscriber](&Config{})
		var suber1 = em.NewSubscriber()
		em.Subscribe(suber1, "topic1", func(subscriber Int64Subscriber, msg any) {
			time.Sleep(100 * time.Millisecond)
		})
		em.Subscribe(suber1, "topic2", func(subscriber Int64Subscriber, msg any) {
			time.Sleep(100 * time.Millisecond)
		})
		em.Subscribe(suber1, "topic3", func(subscriber Int64Subscriber, msg any) {
			time.Sleep(100 * time.Millisecond)
		})
		assert.ElementsMatch(t, em.GetTopicsBySubscriber(suber1), []string{"topic1", "topic2", "topic3"})
		em.UnSubscribe(suber1, "topic1")
		assert.ElementsMatch(t, em.GetTopicsBySubscriber(suber1), []string{"topic2", "topic3"})
		em.UnSubscribeAll(suber1)
		assert.Zero(t, len(em.GetTopicsBySubscriber(suber1)))

		em.UnSubscribeAll(suber1)
		assert.Zero(t, em.CountSubscriberByTopic("topic0"))
	})

	t.Run("", func(t *testing.T) {
		var em = New[Int64Subscriber](nil)
		em.Subscribe(1, "chat", func(subscriber Int64Subscriber, msg any) {})
		em.Subscribe(2, "chat", func(subscriber Int64Subscriber, msg any) {})
		em.UnSubscribe(1, "chat")
		_, exists := em.getBucket(1).Subscribers[1]
		assert.False(t, exists)
	})
}
