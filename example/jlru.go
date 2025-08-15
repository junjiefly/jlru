package example

import (
	jlru "github.com/junjiefly/jlru/lru"
)

type jkv struct {
	lru *jlru.LRU[string, []byte]
}

func newJkv(capacity uint32, maxPriority int, onEvicted func(key string, value []byte) bool) (*jkv, error) {
	kv := &jkv{}
	lru, err := jlru.NewPriorityLRU[string, []byte](int(capacity), byte(maxPriority), jlru.HashXXHASH, onEvicted)
	if err != nil {
		return nil, err
	}
	kv.lru = lru
	return kv, nil
}

func (kv *jkv) Get(key string) ([]byte, bool) {
	data, ok, err := kv.lru.Get(key)
	if err != nil {
		return nil, false
	}
	if !ok {
		return nil, false
	}
	return data, true
}

func (kv *jkv) Has(key string) ([]byte, bool) {
	data, ok, err := kv.lru.Has(key)
	if err != nil {
		return nil, false
	}
	if !ok {
		return nil, false
	}
	return data, true
}

func (kv *jkv) Set(key string, data []byte) bool {
	err := kv.lru.Add(key, data, 0)
	if err == nil {
		return true
	}
	return false
}

func (kv *jkv) SetPriority(key string, data []byte, priority byte) bool {
	err := kv.lru.Add(key, data, priority)
	if err == nil {
		return true
	}
	return false
}

func (kv *jkv) Delete(key string) bool {
	_, ok, err := kv.lru.Remove(key)
	if err != nil {
		return false
	}
	return ok
}

func (kv *jkv) RemoveOldest() bool {
	return kv.lru.RemoveOldest()
}

func (kv *jkv) Length() int {
	return int(kv.lru.Len())
}

func (kv *jkv) Metrics() (uint64, uint64, float64) {
	length := kv.lru.Len()
	metrics := kv.lru.Metrics()
	total := metrics.Hits + metrics.Misses
	if total == 0 {
		return uint64(length), metrics.Evictions, 0
	} else {
		return uint64(length), metrics.Evictions, float64(metrics.Hits) * 100.0 / float64(total)
	}
}
