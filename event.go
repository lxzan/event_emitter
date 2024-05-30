package event_emitter

import (
	"github.com/lxzan/event_emitter/internal/treemap"
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

	// 每个分片里主题表的初始化容量, 根据主题订阅量估算, 默认为0.
	// The initialization capacity of the topic table in each slice is estimated based on the topic subscriptions and is 0 by default.
	BucketSize int64

	// 订阅主题分隔符, 默认为点
	// Subscription subject separator, defaults to a dot.
	Separator string
}

func (c *Config) init() {
	if c.BucketNum <= 0 {
		c.BucketNum = 16
	}
	if c.BucketSize <= 0 {
		c.BucketSize = 0
	}
	if c.Separator == "" {
		c.Separator = "."
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
			Topics: treemap.New[*topicField[T]](conf.Separator),
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

// Publish 发布消息到指定主题, 支持通配符匹配.
// 注意: 如果主题包含有分隔符, 会扫描所有分片.
// Post a message to the topic, support wildcard matching.
// Note: If the topic contains a separator, all slices will be scanned.
func (c *EventEmitter[T]) Publish(topic string, msg any) {
	if strings.Contains(topic, c.conf.Separator) {
		for _, b := range c.buckets {
			b.publish(topic, msg)
		}
	} else {
		c.getBucket(topic).publish(topic, msg)
	}
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
	Topics *treemap.TreeMap[*topicField[T]]
}

// 新增订阅
func (c *bucket[T]) subscribe(suber T, topic string, f eventCallback[T]) {
	c.Lock()
	defer c.Unlock()

	subId := suber.GetSubscriberID()
	ele := topicElement[T]{suber: suber, cb: f}

	if t, ok := c.Topics.Get(topic); ok {
		t.subers[subId] = ele
	} else {
		t = &topicField[T]{subers: make(map[int64]topicElement[T], c.Size)}
		t.subers[subId] = ele
		c.Topics.Put(topic, t)
	}
}

func (c *bucket[T]) publish(topic string, msg any) {
	c.Lock()
	defer c.Unlock()

	c.Topics.Match(topic, func(_ string, t *topicField[T]) {
		for _, v := range t.subers {
			v.cb(v.suber, msg)
		}
	})
}

// 取消某个主题的订阅
func (c *bucket[T]) unSubscribe(suber T, topic string) {
	c.Lock()
	defer c.Unlock()

	v, ok := c.Topics.Get(topic)
	if ok {
		delete(v.subers, suber.GetSubscriberID())
	}
}

func (c *bucket[T]) countTopicSubscriber(topic string) int {
	c.Lock()
	defer c.Unlock()

	v, exists := c.Topics.Get(topic)
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
