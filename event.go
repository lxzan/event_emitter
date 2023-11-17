package event_emitter

import (
	"hash/maphash"
	"sync"
	"sync/atomic"
)

type Config struct {
	// 分片数
	// Number of slices
	BucketNum int64

	// 每个分片的初始化容量, 根据订阅量估算, 默认为0.
	// Initialization capacity of each slice, estimated from subscriptions, default 0.
	BucketSize int64
}

func (c *Config) init() {
	if c.BucketNum <= 0 {
		c.BucketNum = 16
	}
	if c.BucketSize <= 0 {
		c.BucketSize = 0
	}
	c.BucketNum = toBinaryNumber(c.BucketNum)
}

type EventEmitter[T Subscriber[T]] struct {
	conf    Config
	seed    maphash.Seed
	serial  atomic.Int64
	buckets []*bucket[T]
}

// New 创建事件发射器实例
// Creating an EventEmitter Instance
func New[T Subscriber[T]](conf *Config) *EventEmitter[T] {
	if conf == nil {
		conf = new(Config)
	}
	conf.init()

	buckets := make([]*bucket[T], 0, conf.BucketNum)
	for i := int64(0); i < conf.BucketNum; i++ {
		buckets = append(buckets, &bucket[T]{
			Mutex:       sync.Mutex{},
			Topics:      make(map[string]*topicField[T]),
			Subscribers: make(map[int64]*subscriberField[T], conf.BucketSize),
		})
	}

	return &EventEmitter[T]{
		conf:    *conf,
		seed:    maphash.MakeSeed(),
		buckets: buckets,
	}
}

// NewSubscriber 生成订阅ID. 也可以使用自己的ID, 保证唯一即可.
// Generate a subscription ID. You can also use your own ID, just make sure it's unique.
func (c *EventEmitter[T]) NewSubscriber() Int64Subscriber {
	return Int64Subscriber(c.serial.Add(1))
}

func (c *EventEmitter[T]) getBucket(suber T) *bucket[T] {
	i := suber.GetSubscribeID() & (c.conf.BucketNum - 1)
	return c.buckets[i]
}

// Publish 向主题发布消息
// Publish a message to the topic
func (c *EventEmitter[T]) Publish(topic string, msg any) {
	for _, b := range c.buckets {
		b.publish(topic, msg)
	}
}

// Subscribe 订阅主题消息. 注意: 回调函数必须是非阻塞的.
// Subscribe messages from the topic. Note: Callback functions must be non-blocking.
func (c *EventEmitter[T]) Subscribe(suber T, topic string, f func(subscriber T, msg any)) {
	c.getBucket(suber).subscribe(suber, topic, f)
}

// UnSubscribe 取消一个订阅主题
// Cancel a subscribed topic
func (c *EventEmitter[T]) UnSubscribe(suber T, topic string) {
	c.getBucket(suber).unSubscribe(suber, topic)
}

// UnSubscribeAll 取消所有订阅主题
// Cancel all subscribed topics
func (c *EventEmitter[T]) UnSubscribeAll(suber T) {
	c.getBucket(suber).unSubscribeAll(suber)
}

// GetTopicsBySubscriber 通过订阅者获取主题列表
// Get a list of topics by subscriber
func (c *EventEmitter[T]) GetTopicsBySubscriber(suber T) []string {
	return c.getBucket(suber).getSubscriberTopics(suber)
}

// CountSubscriberByTopic 获取主题订阅人数
// Get the number of subscribers to a topic
func (c *EventEmitter[T]) CountSubscriberByTopic(topic string) int {
	var sum = 0
	for _, b := range c.buckets {
		sum += b.countTopicSubscriber(topic)
	}
	return sum
}

type bucket[T Subscriber[T]] struct {
	sync.Mutex
	Topics      map[string]*topicField[T]
	Subscribers map[int64]*subscriberField[T]
}

// 新增订阅
func (c *bucket[T]) subscribe(suber T, topic string, f eventCallback[T]) {
	c.Lock()
	defer c.Unlock()

	// 更新订阅
	subId := suber.GetSubscribeID()
	sub, ok := c.Subscribers[subId]
	if !ok {
		sub = &subscriberField[T]{topics: make(map[string]eventCallback[T])}
	}
	sub.suber = suber
	sub.topics[topic] = f
	c.Subscribers[subId] = sub

	// 更新主题
	t, ok := c.Topics[topic]
	if !ok {
		t = &topicField[T]{subs: make(map[int64]*subscriberField[T])}
	}
	t.subs[subId] = sub
	c.Topics[topic] = t
}

func (c *bucket[T]) publish(topic string, msg any) {
	c.Lock()
	defer c.Unlock()

	s, ok := c.Topics[topic]
	if !ok {
		return
	}
	for _, v := range s.subs {
		if cb, exist := v.topics[topic]; exist {
			cb(v.suber, msg)
		}
	}
}

// 取消某个主题的订阅
func (c *bucket[T]) unSubscribe(suber T, topic string) {
	c.Lock()
	defer c.Unlock()

	subId := suber.GetSubscribeID()
	v1, ok1 := c.Subscribers[subId]
	if ok1 {
		delete(v1.topics, topic)
		if len(v1.topics) == 0 {
			delete(c.Subscribers, subId)
		}

		if v2, ok2 := c.Topics[topic]; ok2 {
			delete(v2.subs, subId)
		}
	}
}

// 取消所有主题的订阅
func (c *bucket[T]) unSubscribeAll(suber T) {
	c.Lock()
	defer c.Unlock()

	subId := suber.GetSubscribeID()
	v1, ok1 := c.Subscribers[subId]
	if ok1 {
		for topic, _ := range v1.topics {
			if v2, ok2 := c.Topics[topic]; ok2 {
				delete(v2.subs, subId)
			}
		}
		delete(c.Subscribers, subId)
	}
}

func (c *bucket[T]) getSubscriberTopics(suber T) []string {
	c.Lock()
	defer c.Unlock()

	v, exists := c.Subscribers[suber.GetSubscribeID()]
	if !exists {
		return nil
	}
	var topics = make([]string, 0, len(v.topics))
	for k, _ := range v.topics {
		topics = append(topics, k)
	}
	return topics
}

func (c *bucket[T]) countTopicSubscriber(topic string) int {
	c.Lock()
	defer c.Unlock()

	v, exists := c.Topics[topic]
	if !exists {
		return 0
	}
	return len(v.subs)
}

func toBinaryNumber(n int64) int64 {
	var x int64 = 1
	for x < n {
		x *= 2
	}
	return x
}
