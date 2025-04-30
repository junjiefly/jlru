package jlru

import (
	"github.com/cespare/xxhash/v2"
	"github.com/junjiefly/jlist"
	"sync"
	"sync/atomic"
)

type ListMetrics struct {
	Inserts   uint64
	Evictions uint64
	Removals  uint64
	Hits      uint64
	Misses    uint64
	Conflict  uint64
	Errors    uint64
}

func hashXXHASH(s string) uint64 {
	return xxhash.Sum64String(s)
}

type LRU struct {
	metrics ListMetrics
	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key string, value []byte) bool

	ll    *jlist.List[string, []byte]
	cache map[uint64]int

	sync.RWMutex
}

func NewLRU(capacity int, onEvicted func(key string, value []byte) bool) *LRU {
	lru := &LRU{
		OnEvicted: onEvicted,
		cache:     make(map[uint64]int, capacity),
	}
	lru.ll = jlist.NewList[string, []byte](capacity)
	return lru
}

// Add adds a value to the cache.
func (lru *LRU) Add(key string, value []byte) error {
	hashId := hashXXHASH(key)
	lru.Lock()
	defer lru.Unlock()
	if idx, ok := lru.cache[hashId]; ok {
		e, err := lru.ll.Entry(idx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return err
		}
		err = lru.ll.MoveToFront(e)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return err
		}
		e.Key = key
		e.HashId = hashId
		e.Value = value
		err = lru.ll.UpdateEntry(idx, e)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return err
		}
		atomic.AddUint64(&lru.metrics.Inserts, 1)
		return nil
	}
	if lru.ll.Len() >= lru.ll.Cap() {
		lru.removeOldest()
	}
	ele, err := lru.ll.PushFront(key, value)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return err
	}
	ele.HashId = hashId
	lru.cache[hashId] = ele.Idx()
	atomic.AddUint64(&lru.metrics.Inserts, 1)
	return nil
}

func (lru *LRU) AddToBack(key string, value []byte) error {
	hashId := hashXXHASH(key)
	lru.Lock()
	defer lru.Unlock()
	if idx, ok := lru.cache[hashId]; ok {
		e, err := lru.ll.Entry(idx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return err
		}
		err = lru.ll.MoveToBack(e)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return err
		}
		e.HashId = hashId
		e.Key = key
		e.Value = value
		err = lru.ll.UpdateEntry(idx, e)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return err
		}
		atomic.AddUint64(&lru.metrics.Inserts, 1)
		return nil
	}
	if lru.ll.Len() >= lru.ll.Cap() {
		lru.removeOldest()
	}
	e, err := lru.ll.PushBack(key, value)
	if err != nil {
		return err
	}
	e.HashId = hashId
	lru.cache[hashId] = e.Idx()
	atomic.AddUint64(&lru.metrics.Inserts, 1)
	return nil
}

// Get looks up a key's value from the cache.
func (lru *LRU) Get(key string) (value []byte, ok bool, err error) {
	hashId := hashXXHASH(key)
	lru.Lock()
	defer lru.Unlock()
	if idx, ok := lru.cache[hashId]; ok {
		e, err := lru.ll.Entry(idx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return value, false, err
		}
		if e.Key != key {
			atomic.AddUint64(&lru.metrics.Conflict, 1)
			return value, false, nil
		}
		atomic.AddUint64(&lru.metrics.Hits, 1)
		value = e.Value
		err = lru.ll.MoveToFront(e)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return value, true, err
		}
		return value, true, nil
	}
	atomic.AddUint64(&lru.metrics.Misses, 1)
	return value, false, nil
}

// Has looks up a key's value from the cache.
func (lru *LRU) Has(key string) (value []byte, ok bool, e error) {
	hashId := hashXXHASH(key)
	lru.RLock()
	defer lru.RUnlock()
	if idx, ok := lru.cache[hashId]; ok {
		e, err := lru.ll.Entry(idx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return value, true, err
		}
		if e.Key == key {
			return e.Value, true, nil
		}
		atomic.AddUint64(&lru.metrics.Conflict, 1)
		return value, false, nil
	}
	return value, false, nil
}

// Remove removes the provided key from the cache.
func (lru *LRU) Remove(key string) (value []byte, ok bool, err error) {
	hashId := hashXXHASH(key)
	lru.Lock()
	defer lru.Unlock()
	idx, ok := lru.cache[hashId]
	if !ok {
		return nil, false, nil
	}
	e, err := lru.ll.Entry(idx)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return nil, false, err
	}
	if e.Key != key {
		atomic.AddUint64(&lru.metrics.Conflict, 1)
		return nil, false, nil
	}
	value = e.Value
	err = lru.removeElement(&e, false)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return value, false, err
	}
	atomic.AddUint64(&lru.metrics.Removals, 1)
	return value, true, nil
}

func (lru *LRU) removeOldest() {
	ele := lru.ll.Back()
	if ele != nil {
		if lru.removeElement(ele, true) == nil {
			atomic.AddUint64(&lru.metrics.Evictions, 1)
		}
	}
	return
}

// RemoveOldest removes the oldest item from the cache.
func (lru *LRU) RemoveOldest() bool {
	lru.Lock()
	defer lru.Unlock()
	ele := lru.ll.Back()
	if ele != nil {
		if lru.removeElement(ele, true) == nil {
			atomic.AddUint64(&lru.metrics.Evictions, 1)
			return true
		}
		atomic.AddUint64(&lru.metrics.Removals, 1)
		return true
	}
	return false
}

func (lru *LRU) Oldest() *jlist.Entry[string, []byte] {
	lru.RLock()
	defer lru.RUnlock()
	e := lru.ll.Back()
	if e == nil {
		return nil
	}
	return e
}

func (lru *LRU) removeElement(e *jlist.Entry[string, []byte], evict bool) error {
	if e == nil {
		return nil
	}
	if evict && lru.OnEvicted != nil {
		if lru.OnEvicted(e.Key, e.Value) {
			_, err := lru.ll.Remove(e)
			if err == nil {
				delete(lru.cache, e.HashId)
			}
			return err
		}
		return nil
	}
	_, err := lru.ll.Remove(e)
	if err != nil {
		return err
	}
	delete(lru.cache, e.HashId)
	return nil
}

// Len returns the number of items in the cache.
func (lru *LRU) Len() int {
	lru.RLock()
	defer lru.RUnlock()
	if lru.ll == nil {
		return 0
	}
	return lru.ll.Len()
}

func (lru *LRU) Metrics() ListMetrics {
	lru.RLock()
	defer lru.RUnlock()
	return lru.metrics
}

// Clear purges all stored items from the cache.
func (lru *LRU) Clear() {
	for id, idx := range lru.cache {
		e, err := lru.ll.Entry(idx)
		if err == nil {
			if lru.OnEvicted != nil {
				if !lru.OnEvicted(e.Key, e.Value) {
					continue
				}
			}
			_, _ = lru.ll.Remove(&e)
			delete(lru.cache, id)
		}
	}
	lru.ll.Clear()
	lru.ll = nil
	lru.cache = nil
}
