package event_emitter

import "sync"

type eventCallback func(msg any)

type topicField[T comparable, S Subscriber[T]] struct {
	subers map[T]topicElement[S]
}

type topicElement[S any] struct {
	suber S
	cb    eventCallback
}

type (
	Subscriber[T comparable] interface {
		// GetSubscriberID 获取订阅者唯一ID
		// Get subscriber unique ID
		GetSubscriberID() T

		// GetMetadata 获取元数据
		// Getting Metadata
		GetMetadata() Metadata
	}

	Metadata interface {
		Load(key string) (value any, exist bool)
		Store(key string, value any)
		Delete(key string)
		Range(f func(key string, value any) bool)
	}
)

type subscriber[T comparable] struct {
	id T
	md Metadata
}

func (c *subscriber[T]) GetMetadata() Metadata { return c.md }

func (c *subscriber[T]) GetSubscriberID() T { return c.id }

func newSmap() *smap { return &smap{data: make(map[string]any)} }

type smap struct {
	sync.RWMutex
	data map[string]any
}

func (c *smap) Load(key string) (value any, exist bool) {
	c.RLock()
	defer c.RUnlock()
	value, exist = c.data[key]
	return
}

func (c *smap) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.data, key)
}

func (c *smap) Store(key string, value any) {
	c.Lock()
	defer c.Unlock()
	c.data[key] = value
}

func (c *smap) Range(f func(key string, value any) bool) {
	c.RLock()
	defer c.RUnlock()

	for k, v := range c.data {
		if !f(k, v) {
			return
		}
	}
}
