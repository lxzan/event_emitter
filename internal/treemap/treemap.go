package treemap

import (
	"strings"
)

func New[T any](sep string) *TreeMap[T] {
	return &TreeMap[T]{
		sep:  sep,
		root: &Element[T]{children: make(map[string]*Element[T])},
	}
}

type (
	TreeMap[T any] struct {
		sep  string
		root *Element[T]
	}

	Element[T any] struct {
		ok       bool
		children map[string]*Element[T]
		key      string
		value    T
	}
)

func (c *TreeMap[T]) Put(key string, val T) {
	node := &Element[T]{key: key, value: val, ok: true}
	c.doPut(c.root, node, key)
}

func (c *TreeMap[T]) doPut(far, son *Element[T], key string) {
	if far.children == nil {
		far.children = make(map[string]*Element[T])
	}

	index := strings.Index(key, c.sep)
	if index < 0 {
		if node := far.children[key]; node == nil {
			far.children[key] = son
		} else {
			node.value, node.ok = son.value, son.ok
		}
		return
	}

	s0, s1 := key[:index], key[index+1:]
	next := far.children[s0]
	if far.children[s0] == nil {
		next = &Element[T]{}
		far.children[s0] = next
	}
	c.doPut(next, son, s1)
}

func (c *TreeMap[T]) Get(key string) (v T, exist bool) {
	if result, ok := c.root.children[key]; ok && result.ok {
		v, exist = result.value, true
	}
	return
}

func (c *TreeMap[T]) Match(key string, cb func(k string, v T)) {
	c.doMatch(c.root, key, cb)
}

func (c *TreeMap[T]) doMatch(cur *Element[T], key string, cb func(string, T)) {
	index := strings.Index(key, c.sep)
	if index < 0 {
		if key == "*" {
			for _, node := range cur.children {
				if node.ok {
					cb(node.key, node.value)
				}
			}
		} else {
			if result, ok := cur.children[key]; ok && result.ok {
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
