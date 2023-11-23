# EventEmitter

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/event_emitter/actions/workflows/go.yml/badge.svg

[2]: https://github.com/lxzan/event_emitter/actions/workflows/go.yml

[3]: https://codecov.io/gh/lxzan/event_emitter/graph/badge.svg?token=WnGHinZwVR

[4]: https://codecov.io/gh/lxzan/event_emitter

### Install

```bash
go get -v github.com/lxzan/event_emitter@latest
```

### Quick Start

```go
package main

import (
	"fmt"
	"github.com/lxzan/event_emitter"
)

func main() {
	var em = event_emitter.New[event_emitter.Subscriber[any]](&event_emitter.Config{
		BucketNum:  16,
		BucketSize: 128,
	})

	var suber1 = em.NewSubscriber()
	em.Subscribe(suber1, "greet", func(subscriber event_emitter.Subscriber[any], msg any) {
		fmt.Printf("recv: sub_id=%d, msg=%v\n", subscriber.GetSubscriberID(), msg)
	})
	em.Subscribe(suber1, "greet1", func(subscriber event_emitter.Subscriber[any], msg any) {
		fmt.Printf("recv: sub_id=%d, msg=%v\n", subscriber.GetSubscriberID(), msg)
	})

	var suber2 = em.NewSubscriber()
	em.Subscribe(suber2, "greet1", func(subscriber event_emitter.Subscriber[any], msg any) {
		fmt.Printf("recv: sub_id=%d, msg=%v\n", subscriber.GetSubscriberID(), msg)
	})

	em.Publish("greet1", "hello!")
}
```

### GWS Broadcast

```go
package main

import (
	"github.com/lxzan/event_emitter"
	"github.com/lxzan/gws"
)

type Socket struct{ *gws.Conn }

func (c *Socket) GetSubscriberID() int64 {
	userId, _ := c.Session().Load("userId")
	return userId.(int64)
}

func (c *Socket) GetMetadata() event_emitter.Metadata {
	return c.Conn.Session()
}

func Sub(em *event_emitter.EventEmitter[*Socket], topic string, socket *Socket) {
	em.Subscribe(socket, topic, func(subscriber *Socket, msg any) {
		_ = msg.(*gws.Broadcaster).Broadcast(subscriber.Conn)
	})
}

func Pub(em *event_emitter.EventEmitter[*Socket], topic string, op gws.Opcode, msg []byte) {
	var broadcaster = gws.NewBroadcaster(op, msg)
	defer broadcaster.Close()
	em.Publish(topic, broadcaster)
}
```