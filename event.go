package event_emitter

import (
	"hash/maphash"
	"strings"
	"sync"
	"sync/atomic"
)

const subTopic = "sub-topic-"

type Config struct {
	// 分片数
	// Number of slices
	BucketNum int64

	// 每个分片的初始化容量, 根据主题订阅量估算, 默认为0.
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
	serial  atomic.Int64
	seed    maphash.Seed
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
			Mutex:  sync.Mutex{},
			Size:   conf.BucketSize,
			Topics: make(map[string]*topicField[T]),
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
func (c *EventEmitter[T]) NewSubscriber() Subscriber[any] {
	return &Int64Subscriber{
		id: c.serial.Add(1),
		md: newSmap(),
	}
}

func (c *EventEmitter[T]) getBucket(topic string) *bucket[T] {
	i := maphash.String(c.seed, topic) & uint64(c.conf.BucketNum-1)
	return c.buckets[i]
}

// Publish 向主题发布消息
// Publish a message to the topic
func (c *EventEmitter[T]) Publish(topic string, msg any) {
	c.getBucket(topic).publish(topic, msg)
}

// Subscribe 订阅主题消息. 注意: 回调函数必须是非阻塞的.
// Subscribe messages from the topic. Note: Callback functions must be non-blocking.
func (c *EventEmitter[T]) Subscribe(suber T, topic string, f func(subscriber T, msg any)) {
	suber.GetMetadata().Store(subTopic+topic, topic)
	c.getBucket(topic).subscribe(suber, topic, f)
}

// UnSubscribe 取消订阅一个主题
// Cancel a subscribed topic
func (c *EventEmitter[T]) UnSubscribe(suber T, topic string) {
	suber.GetMetadata().Delete(subTopic + topic)
	c.getBucket(topic).unSubscribe(suber, topic)
}

// UnSubscribeAll 取消订阅所有主题
// Cancel all subscribed topics
func (c *EventEmitter[T]) UnSubscribeAll(suber T) {
	var topics []string
	var md = suber.GetMetadata()
	md.Range(func(key string, value any) bool {
		if strings.HasPrefix(key, subTopic) {
			topics = append(topics, value.(string))
		}
		return true
	})
	for _, topic := range topics {
		md.Delete(subTopic + topic)
		c.getBucket(topic).unSubscribe(suber, topic)
	}
}

// GetTopicsBySubscriber 通过订阅者获取主题列表
// Get a list of topics by subscriber
func (c *EventEmitter[T]) GetTopicsBySubscriber(suber T) []string {
	var topics []string
	suber.GetMetadata().Range(func(key string, value any) bool {
		if strings.HasPrefix(key, subTopic) {
			topics = append(topics, value.(string))
		}
		return true
	})
	return topics
}

// CountSubscriberByTopic 获取主题订阅人数
// Get the number of subscribers to a topic
func (c *EventEmitter[T]) CountSubscriberByTopic(topic string) int {
	return c.getBucket(topic).countTopicSubscriber(topic)
}

type bucket[T Subscriber[T]] struct {
	sync.Mutex
	Size   int64
	Topics map[string]*topicField[T]
}

// 新增订阅
func (c *bucket[T]) subscribe(suber T, topic string, f eventCallback[T]) {
	c.Lock()
	defer c.Unlock()

	subId := suber.GetSubscriberID()
	ele := topicElement[T]{suber: suber, cb: f}

	t, ok := c.Topics[topic]
	if !ok {
		t = &topicField[T]{subers: make(map[int64]topicElement[T], c.Size)}
		t.subers[subId] = ele
		c.Topics[topic] = t
		return
	}

	t.subers[subId] = ele
}

func (c *bucket[T]) publish(topic string, msg any) {
	c.Lock()
	defer c.Unlock()

	t, ok := c.Topics[topic]
	if !ok {
		return
	}
	for _, v := range t.subers {
		v.cb(v.suber, msg)
	}
}

// 取消某个主题的订阅
func (c *bucket[T]) unSubscribe(suber T, topic string) {
	c.Lock()
	defer c.Unlock()

	v, ok := c.Topics[topic]
	if ok {
		delete(v.subers, suber.GetSubscriberID())
	}
}

func (c *bucket[T]) countTopicSubscriber(topic string) int {
	c.Lock()
	defer c.Unlock()

	v, exists := c.Topics[topic]
	if !exists {
		return 0
	}
	return len(v.subers)
}

func toBinaryNumber(n int64) int64 {
	var x int64 = 1
	for x < n {
		x *= 2
	}
	return x
}
