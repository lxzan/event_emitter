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
	"context"
	"fmt"
	"github.com/lxzan/event_emitter"
	"time"
)

func main() {
	var em = event_emitter.New(&event_emitter.Config{
		BucketNum:   16,
		BucketCap:   128,
		Concurrency: 8,
	})
	em.Subscribe(em.NewSubscriber(), "greet", func(msg any) {
		fmt.Printf("recv: %v\n", msg)
	})
	em.Subscribe(em.NewSubscriber(), "greet", func(msg any) {
		fmt.Printf("recv: %v\n", msg)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = em.Publish(ctx, "greet", "hello!")
	time.Sleep(time.Second)
}
```