package event_emitter

import "sync"

type eventCallback[T Subscriber[T]] func(suber T, msg any)

type topicField[T Subscriber[T]] struct {
	subers map[int64]topicElement[T]
}

type topicElement[T Subscriber[T]] struct {
	suber T
	cb    eventCallback[T]
}

type (
	Subscriber[T any] interface {
		GetSubscriberID() int64 // 获取订阅者唯一ID
		GetMetadata() Metadata
	}

	Metadata interface {
		Load(key string) (value any, exist bool)
		Store(key string, value any)
		Delete(key string)
		Range(f func(key string, value any) bool)
	}
)

type Int64Subscriber struct {
	id int64
	md Metadata
}

func (c *Int64Subscriber) GetMetadata() Metadata {
	return c.md
}

func (c *Int64Subscriber) GetSubscriberID() int64 {
	return c.id
}

type smap struct{ sync.Map }

func (c *smap) Delete(key string) {
	c.Map.Delete(key)
}

func (c *smap) Range(f func(key string, value any) bool) {
	c.Map.Range(func(k, v any) bool { return f(k.(string), v) })
}

func (c *smap) Load(key string) (value any, exist bool) {
	return c.Map.Load(key)
}

func (c *smap) Store(key string, value any) {
	c.Map.Store(key, value)
}

func NewMetadata() Metadata { return new(smap) }
