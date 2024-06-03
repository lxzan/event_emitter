package event_emitter

import (
	"hash/maphash"
	"strings"
	"sync"
)

const subTopic = "sub-topic-"

type Config struct {
	// 分片数
	// Number of slices
	BucketNum int64

	// 每个分片里主题表的初始化容量, 根据主题订阅量估算, 默认为0.
	// The initialization capacity of the topic table in each slice is estimated based on the topic subscriptions and is 0 by default.
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

// EventEmitter T: 订阅ID模板类型 S: 订阅者模板类型
// T:subscription id template type S:subscriber template type
type EventEmitter[T comparable, S Subscriber[T]] struct {
	conf    Config
	seed    maphash.Seed
	buckets []*bucket[T, S]
}

// New 创建事件发射器实例
// Creating an EventEmitter Instance
func New[T comparable, S Subscriber[T]](conf *Config) *EventEmitter[T, S] {
	if conf == nil {
		conf = new(Config)
	}
	conf.init()

	buckets := make([]*bucket[T, S], 0, conf.BucketNum)
	for i := int64(0); i < conf.BucketNum; i++ {
		buckets = append(buckets, &bucket[T, S]{
			Mutex:  sync.Mutex{},
			Size:   conf.BucketSize,
			Topics: make(map[string]*topicField[T, S]),
		})
	}

	return &EventEmitter[T, S]{
		conf:    *conf,
		seed:    maphash.MakeSeed(),
		buckets: buckets,
	}
}

// NewSubscriber 创建一个订阅者, 订阅Id需要保证唯一.
// Create a subscriber, the subscription id needs to be unique.
func (c *EventEmitter[T, S]) NewSubscriber(id T) Subscriber[T] {
	return &subscriber[T]{id: id, md: newSmap()}
}

func (c *EventEmitter[T, S]) getBucket(topic string) *bucket[T, S] {
	i := maphash.String(c.seed, topic) & uint64(c.conf.BucketNum-1)
	return c.buckets[i]
}

// Publish 发布消息到指定主题
// Post a message to the topic.
func (c *EventEmitter[T, S]) Publish(topic string, msg any) {
	c.getBucket(topic).publish(topic, msg)
}

// Subscribe 订阅主题消息. 注意: 回调函数必须是非阻塞的.
// Subscribe messages from the topic. Note: Callback functions must be non-blocking.
func (c *EventEmitter[T, S]) Subscribe(suber S, topic string, f func(msg any)) {
	suber.GetMetadata().Store(subTopic+topic, topic)
	c.getBucket(topic).subscribe(suber, topic, f)
}

// UnSubscribe 取消订阅一个主题
// Cancel a subscribed topic
func (c *EventEmitter[T, S]) UnSubscribe(suber S, topic string) {
	suber.GetMetadata().Delete(subTopic + topic)
	c.getBucket(topic).unSubscribe(suber, topic)
}

// UnSubscribeAll 取消订阅所有主题
// Cancel all subscribed topics
func (c *EventEmitter[T, S]) UnSubscribeAll(suber S) {
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
func (c *EventEmitter[T, S]) GetTopicsBySubscriber(suber S) []string {
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
func (c *EventEmitter[T, S]) CountSubscriberByTopic(topic string) int {
	return c.getBucket(topic).countTopicSubscriber(topic)
}

type bucket[T comparable, S Subscriber[T]] struct {
	sync.Mutex
	Size   int64
	Topics map[string]*topicField[T, S]
}

// 新增订阅
func (c *bucket[T, S]) subscribe(suber S, topic string, f eventCallback) {
	c.Lock()
	defer c.Unlock()

	subId := suber.GetSubscriberID()
	ele := topicElement[S]{suber: suber, cb: f}

	if t, ok := c.Topics[topic]; ok {
		t.subers[subId] = ele
	} else {
		t = &topicField[T, S]{subers: make(map[T]topicElement[S], c.Size)}
		t.subers[subId] = ele
		c.Topics[topic] = t
	}
}

func (c *bucket[T, S]) publish(topic string, msg any) {
	c.Lock()
	defer c.Unlock()

	if t, ok := c.Topics[topic]; ok {
		for _, v := range t.subers {
			v.cb(msg)
		}
	}
}

// 取消某个主题的订阅
func (c *bucket[T, S]) unSubscribe(suber S, topic string) {
	c.Lock()
	defer c.Unlock()

	if v, ok := c.Topics[topic]; ok {
		delete(v.subers, suber.GetSubscriberID())
	}
}

func (c *bucket[T, S]) countTopicSubscriber(topic string) int {
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
