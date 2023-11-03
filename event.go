package event_emitter

import (
	"context"
	"hash/maphash"
	"sync"
	"sync/atomic"
)

type Config struct {
	BucketNum   int64 // 分片数 (number of slices)
	BucketCap   int64 // 每个分片的初始化容量 (initialized capacity per slice)
	Concurrency int64 // 每个主题的并发度 (Concurrency of each topic)
}

func (c *Config) init() {
	if c.BucketNum <= 0 {
		c.BucketNum = 16
	}
	if c.BucketCap <= 0 {
		c.BucketCap = 0
	}
	if c.Concurrency <= 0 {
		c.Concurrency = 16
	}
	c.BucketNum = toBinaryNumber(c.BucketNum)
}

type EventEmitter struct {
	conf    Config
	seed    maphash.Seed
	serial  atomic.Int64
	buckets []*bucket
}

// New 创建事件发射器实例
// Creating an EventEmitter Instance
func New(conf *Config) *EventEmitter {
	if conf == nil {
		conf = new(Config)
	}
	conf.init()

	buckets := make([]*bucket, 0, conf.BucketNum)
	for i := int64(0); i < conf.BucketNum; i++ {
		buckets = append(buckets, &bucket{
			Mutex:       sync.Mutex{},
			Topics:      make(map[string]*topicField),
			Subscribers: make(map[int64]*subscriberField, conf.BucketCap),
		})
	}

	return &EventEmitter{
		conf:    *conf,
		seed:    maphash.MakeSeed(),
		buckets: buckets,
	}
}

// NewSubscriber 获取订阅号
// Get subscription number
func (c *EventEmitter) NewSubscriber() (subId int64) {
	return c.serial.Add(1)
}

func (c *EventEmitter) getBucketByTopic(topic string) *bucket {
	i := maphash.String(c.seed, topic) & uint64(c.conf.BucketNum-1)
	return c.buckets[i]
}

func (c *EventEmitter) getBucketBySubId(subId int64) *bucket {
	i := subId & (c.conf.BucketNum - 1)
	return c.buckets[i]
}

// Publish 向主题发布消息
// Publish a message to the topic
func (c *EventEmitter) Publish(ctx context.Context, topic string, msg any) error {
	t, ok := c.getBucketByTopic(topic).getTopic(topic)
	if !ok {
		return nil
	}

	t.Lock()
	defer t.Unlock()

	for _, v := range t.subs {
		if err := t.Emit(ctx, msg, v.cb); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe 订阅主题消息
// Subscribe messages from the topic
func (c *EventEmitter) Subscribe(subId int64, topic string, f func(msg any)) {
	sub := c.getBucketBySubId(subId).addSubscriber(subId, topic, f)
	c.getBucketByTopic(topic).addTopic(topic, sub, c.conf.Concurrency)
}

// UnSubscribe 取消一个订阅主题
// Cancel a subscribed topic
func (c *EventEmitter) UnSubscribe(subId int64, topic string) {
	if t, ok := c.getBucketByTopic(topic).getTopic(topic); ok {
		t.Delete(subId)
	}

	b := c.getBucketBySubId(subId)
	if s, ok := b.getSubscriber(subId); ok {
		if s.Delete(topic) == 0 {
			b.Lock()
			delete(b.Subscribers, subId)
			b.Unlock()
		}
	}
}

// UnSubscribeAll 取消所有订阅主题
// Cancel all subscribed topics
func (c *EventEmitter) UnSubscribeAll(subId int64) {
	s, ok := c.getBucketBySubId(subId).deleteSubscriber(subId)
	if !ok {
		return
	}
	for k, _ := range s.topics {
		c.UnSubscribe(subId, k)
	}
}

// GetTopicsBySubId 通过订阅号获取主题列表
// Get a list of topics by subscription
func (c *EventEmitter) GetTopicsBySubId(subId int64) []string {
	return c.getBucketBySubId(subId).getSubscriberTopics(subId)
}

// CountSubscriberByTopic 获取主题订阅人数
// Get the number of subscribers to a topic
func (c *EventEmitter) CountSubscriberByTopic(topic string) int {
	return c.getBucketByTopic(topic).countTopicSubscriber(topic)
}

type bucket struct {
	sync.Mutex
	Topics      map[string]*topicField
	Subscribers map[int64]*subscriberField
}

func (c *bucket) addSubscriber(subId int64, topic string, f func(msg any)) *subscriberField {
	c.Lock()
	defer c.Unlock()

	sub, ok := c.Subscribers[subId]
	if !ok {
		sub = &subscriberField{
			subId:  subId,
			cb:     f,
			topics: make(map[string]struct{}),
		}
		c.Subscribers[subId] = sub
	}
	sub.Add(topic)
	return sub
}

func (c *bucket) addTopic(topic string, sub *subscriberField, concurrency int64) {
	c.Lock()
	defer c.Unlock()

	t, ok := c.Topics[topic]
	if !ok {
		t = &topicField{
			channel: make(chan struct{}, concurrency),
			subs:    make(map[int64]*subscriberField),
		}
		c.Topics[topic] = t
	}
	t.Add(sub.subId, sub)
}

func (c *bucket) getTopic(topic string) (*topicField, bool) {
	c.Lock()
	defer c.Unlock()
	v, exists := c.Topics[topic]
	return v, exists
}

func (c *bucket) getSubscriber(sudId int64) (*subscriberField, bool) {
	c.Lock()
	defer c.Unlock()
	v, exists := c.Subscribers[sudId]
	return v, exists
}

func (c *bucket) getSubscriberTopics(sudId int64) []string {
	c.Lock()
	defer c.Unlock()
	v, exists := c.Subscribers[sudId]
	if !exists {
		return nil
	}

	var topics = make([]string, 0, len(v.topics))
	v.Lock()
	for k, _ := range v.topics {
		topics = append(topics, k)
	}
	v.Unlock()
	return topics
}

func (c *bucket) countTopicSubscriber(topic string) int {
	c.Lock()
	defer c.Unlock()
	v, exists := c.Topics[topic]
	if !exists {
		return 0
	}

	v.Lock()
	n := len(v.subs)
	v.Unlock()
	return n
}

func (c *bucket) deleteSubscriber(sudId int64) (*subscriberField, bool) {
	c.Lock()
	defer c.Unlock()
	v, exists := c.Subscribers[sudId]
	if !exists {
		return nil, false
	}
	delete(c.Subscribers, sudId)
	return v, true
}

func toBinaryNumber(n int64) int64 {
	var x int64 = 1
	for x < n {
		x *= 2
	}
	return x
}
