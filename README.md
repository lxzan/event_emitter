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

### Usage

```go
package main

import (
	"fmt"
	"github.com/lxzan/event_emitter"
)

func main() {
	var em = event_emitter.New[event_emitter.Int64Subscriber](&event_emitter.Config{
		BucketNum:  16,
		BucketSize: 128,
	})

	var suber1 = em.NewSubscriber()
	em.Subscribe(suber1, "greet", func(subscriber event_emitter.Int64Subscriber, msg any) {
		fmt.Printf("recv0: %v\n", msg)
	})
	em.Subscribe(suber1, "greet1", func(subscriber event_emitter.Int64Subscriber, msg any) {
		fmt.Printf("recv1: %v\n", msg)
	})

	var suber2 = em.NewSubscriber()
	em.Subscribe(suber2, "greet1", func(subscriber event_emitter.Int64Subscriber, msg any) {
		fmt.Printf("recv2: %v\n", msg)
	})

	em.Publish("greet1", "hello!")
}
```