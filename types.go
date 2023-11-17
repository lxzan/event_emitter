package event_emitter

type eventCallback[T Subscriber[T]] func(suber T, msg any)

type topicField[T Subscriber[T]] struct {
	subs map[int64]*subscriberField[T]
}

type subscriberField[T Subscriber[T]] struct {
	suber  T
	topics map[string]eventCallback[T]
}

type Subscriber[T any] interface {
	GetSubscribeID() int64
}

type Int64Subscriber int64

func (c Int64Subscriber) GetSubscribeID() int64 {
	return int64(c)
}
