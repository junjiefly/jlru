package example

import (
	"github.com/cespare/xxhash/v2"
	"github.com/elastic/go-freelru"
)

func hashXXHASH(s string) uint32 {
	return uint32(xxhash.Sum64String(s))
}

func newFreeKV(capacity uint32, onEvicted func(key string, value []byte)) (*freeKV, error) {
	kv := &freeKV{}
	lru, err := freelru.NewSynced[string, []byte](capacity, hashXXHASH)
	//lru, err := util.New[string, []byte](capacity, hashXXHASH)
	if err != nil {
		return nil, err
	}
	kv.lru = lru
	kv.lru.SetOnEvict(onEvicted)
	return kv, nil
}

type freeKV struct {
	lru *freelru.SyncedLRU[string, []byte] //进行淘汰管理
	//lru *util.LRU[string, []byte] //进行淘汰管理
}

func (kv *freeKV) Get(key string) ([]byte, bool) {
	data, ok := kv.lru.Get(key)
	if !ok {
		return nil, false
	}
	return data, true
}

func (kv *freeKV) Has(key string) ([]byte, bool) {
	data, ok := kv.lru.Peek(key)
	if !ok {
		return nil, false
	}
	return data, true
}

func (kv *freeKV) Set(key string, data []byte) bool {
	ok := kv.lru.Add(key, data)
	return ok
}

func (kv *freeKV) SetPriority(key string, data []byte, priority byte) bool {
	ok := kv.lru.Add(key, data)
	return ok
}

func (kv *freeKV) Delete(key string) bool {
	ok := kv.lru.Remove(key)
	return ok
}

func (kv *freeKV) RemoveOldest() bool {
	_, _, ok := kv.lru.RemoveOldest()
	return ok
}

func (kv *freeKV) Length() int {
	length := kv.lru.Len()
	return length
}

func (kv *freeKV) Metrics() (uint64, uint64, float64) {
	length := uint64(kv.lru.Len())
	metrics := kv.lru.Metrics()
	total := metrics.Hits + metrics.Misses
	if total == 0 {
		return length, metrics.Evictions, 0
	}
	//fmt.Println("kv cnt:", length, "Evictions:", metrics.Evictions, "rate:", float64(metrics.Hits)/float64(total))
	return length, metrics.Evictions, float64(metrics.Hits) * 100.0 / float64(total)
}
