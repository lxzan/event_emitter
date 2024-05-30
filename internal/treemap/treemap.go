package treemap

import (
	"github.com/lxzan/event_emitter/internal/helper"
	"strings"
)

func New[T any](sep string) *TreeMap[T] {
	return &TreeMap[T]{sep: sep, children: make(map[string]*TreeMap[T])}
}

type TreeMap[T any] struct {
	children map[string]*TreeMap[T]
	sep      string
	key      string
	value    T
}

func (c *TreeMap[T]) Put(key string, val T) {
	if !strings.Contains(key, c.sep) {
		c.children[key] = &TreeMap[T]{key: key, value: val}
		return
	}

	var list = helper.Split(key, c.sep)
	if len(list) == 0 {
		return
	}
	c.doPut(c, 0, list, key, val)
}

func (c *TreeMap[T]) doPut(node *TreeMap[T], index int, list []string, key string, val T) {
	var segment = list[index]
	if node.children == nil {
		node.children = make(map[string]*TreeMap[T])
	}

	var next = node.children[segment]
	if node.children[segment] == nil {
		next = &TreeMap[T]{}
		node.children[segment] = next
	}

	if index+1 == len(list) {
		next.key = key
		next.value = val
	} else {
		c.doPut(next, index+1, list, key, val)
	}
}

func (c *TreeMap[T]) Get(key string) (v T, exist bool) {
	if result, ok := c.children[key]; ok {
		v, exist = result.value, true
	}
	return
}

func (c *TreeMap[T]) Match(key string, cb func(k string, v T)) {
	c.doMatch(c, key, cb)
}

func (c *TreeMap[T]) doMatch(cur *TreeMap[T], key string, cb func(string, T)) {
	index := strings.Index(key, c.sep)
	if index < 0 {
		if key == "*" {
			for _, node := range cur.children {
				cb(node.key, node.value)
			}
		} else {
			if result, ok := cur.children[key]; ok {
				cb(result.key, result.value)
			}
		}
		return
	}

	s0, s1 := key[:index], key[index+1:]
	if s0 == "*" {
		for _, node := range cur.children {
			c.doMatch(node, s1, cb)
		}
	} else {
		if v, ok := cur.children[s0]; ok {
			c.doMatch(v, s1, cb)
		}
	}
}
