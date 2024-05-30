package treemap

import (
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
	} else {
		c.doPut(c, &TreeMap[T]{key: key, value: val}, key)
	}
}

func (c *TreeMap[T]) doPut(far, son *TreeMap[T], key string) {
	if far.children == nil {
		far.children = make(map[string]*TreeMap[T])
	}

	index := strings.Index(key, c.sep)
	if index < 0 {
		far.children[key] = son
		return
	}

	s0, s1 := key[:index], key[index+1:]
	next := far.children[s0]
	if far.children[s0] == nil {
		next = &TreeMap[T]{}
		far.children[s0] = next
	}
	c.doPut(next, son, s1)
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
