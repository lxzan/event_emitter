package event_emitter

import (
	"context"
	"sync"
)

type eventCallback func(msg any)

type topicField struct {
	sync.Mutex
	channel chan struct{}
	subs    map[int64]*subscriberField
}

func (c *topicField) Add(k int64, v *subscriberField) {
	c.Lock()
	c.subs[k] = v
	c.Unlock()
}

func (c *topicField) Delete(k int64) {
	c.Lock()
	delete(c.subs, k)
	c.Unlock()
}

func (c *topicField) Emit(ctx context.Context, msg any, f func(any)) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.channel <- struct{}{}:
	}

	go func() {
		f(msg)
		<-c.channel
	}()
	return nil
}

type subscriberField struct {
	sync.Mutex
	subId  int64
	cb     eventCallback
	topics map[string]struct{}
}

func (c *subscriberField) Add(k string) {
	c.Lock()
	c.topics[k] = struct{}{}
	c.Unlock()
}

func (c *subscriberField) Delete(k string) int {
	c.Lock()
	delete(c.topics, k)
	n := len(c.topics)
	c.Unlock()
	return n
}
