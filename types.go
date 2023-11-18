package event_emitter

type eventCallback[T Subscriber[T]] func(suber T, msg any)

type topicField[T Subscriber[T]] struct {
	subers map[int64]*subscriberField[T]
}

type subscriberField[T Subscriber[T]] struct {
	suber  T
	topics map[string]eventCallback[T]
}

type Subscriber[T any] interface {
	GetSubscriberID() int64 // 获取订阅者唯一ID
}

type Int64Subscriber int64

func (c Int64Subscriber) GetSubscriberID() int64 {
	return int64(c)
}
