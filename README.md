# EventEmitter

### Simple, Fast and Thread-Safe Event Pub/Sub Library for Golang

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/event_emitter/actions/workflows/go.yml/badge.svg

[2]: https://github.com/lxzan/event_emitter/actions/workflows/go.yml

[3]: https://codecov.io/gh/lxzan/event_emitter/graph/badge.svg?token=WnGHinZwVR

[4]: https://codecov.io/gh/lxzan/event_emitter

### Introduction

EventEmitter is a simple, fast and thread-safe event Pub/Sub library for Golang.

EventEmitter only supports in memory storage, and does not support persistence, so it is not suitable for scenarios that
require persistence. It can only be used in the application itself, does not support distribution, if you need
distribution, please use RabittMQ, Kafka etc.

EventEmitter designed for scenarios that require high concurrency and high performance.

## Install

```bash
go get -v github.com/lxzan/event_emitter@latest
```

### Quick Start

It very easy to use, just create a event emitter, then subscribe and publish messages. It can be used in any place of
your application.

**The following is a simple example.**

```go
package main

import (
	"fmt"
	"github.com/lxzan/event_emitter"
	"sync/atomic"
)

func main() {
	// create a event emitter
	var em = event_emitter.New[int64, event_emitter.Subscriber[int64]](&event_emitter.Config{
		BucketNum:  16,
		BucketSize: 128,
	})

	var increaser = new(atomic.Int64)

	// create a subscriber
	var suber1 = em.NewSubscriber(increaser.Add(1))

	// subscribe topic "greet"
	em.Subscribe(suber1, "greet", func(msg any) {
		fmt.Printf("recv: sub_id=%d, msg=%v\n", suber1.GetSubscriberID(), msg)
	})
	// subscribe topic "greet1"
	em.Subscribe(suber1, "greet1", func(msg any) {
		fmt.Printf("recv: sub_id=%d, msg=%v\n", suber1.GetSubscriberID(), msg)
	})

	// create another subscriber
	var suber2 = em.NewSubscriber(increaser.Add(1))

	// subscribe topic "greet1"
	em.Subscribe(suber2, "greet1", func(msg any) {
		fmt.Printf("recv: sub_id=%d, msg=%v\n", suber2.GetSubscriberID(), msg)
	})

	// publish message to topic "greet"
	em.Publish("greet1", "hello!")
}

```

### More Examples

#### GWS Broadcast

Use the event_emitter package to implement the publish-subscribe model. Wrap gws.Conn in a structure and implement the
GetSubscriberID method to get the subscription ID, which must be unique. The subscription ID is used to identify the
subscriber, who can only receive messages on the subject of his subscription.

This example is useful for building chat rooms or push messages using gws. This means that a user can subscribe to one
or more topics via websocket, and when a message is posted to that topic, all subscribers will receive the message.

```go
package main

import (
	"github.com/lxzan/event_emitter"
	"github.com/lxzan/gws"
)

type Subscriber gws.Conn

func NewSubscriber(conn *gws.Conn) *Subscriber { return (*Subscriber)(conn) }

func (c *Subscriber) GetSubscriberID() int64 {
	userId, _ := c.GetMetadata().Load("userId")
	return userId.(int64)
}

func (c *Subscriber) GetMetadata() event_emitter.Metadata { return c.Conn().Session() }

func (c *Subscriber) Conn() *gws.Conn { return (*gws.Conn)(c) }

func Subscribe(em *event_emitter.EventEmitter[int64, *Subscriber], s *Subscriber, topic string) {
	em.Subscribe(s, topic, func(msg any) {
		_ = msg.(*gws.Broadcaster).Broadcast(s.Conn())
	})
}

func Publish(em *event_emitter.EventEmitter[int64, *Subscriber], topic string, msg []byte) {
	var broadcaster = gws.NewBroadcaster(gws.OpcodeText, msg)
	defer broadcaster.Close()
	em.Publish(topic, broadcaster)
}

```
