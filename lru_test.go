package jlru

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestAddSingleElement(t *testing.T) {
	lru := NewLRU(2, nil)
	key := "test"
	value := []byte("value")
	err := lru.Add(key, value)
	assert.NoError(t, err)
	assert.Equal(t, 1, lru.Len())
	assert.EqualValues(t, 1, lru.Metrics().Inserts)
}

func TestAddExceedCapacity(t *testing.T) {
	lru := NewLRU(1, nil)
	lru.Add("key1", []byte("value1"))
	lru.Add("key2", []byte("value2"))
	assert.Equal(t, 1, lru.Len())
	assert.EqualValues(t, 1, lru.Metrics().Evictions)
}

func TestGetExisting(t *testing.T) {
	lru := NewLRU(2, nil)
	key := "test"
	value := []byte("value")
	lru.Add(key, value)
	got, ok, err := lru.Get(key)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, bytes.Equal(got, value))
	assert.EqualValues(t, 1, lru.Metrics().Hits)
}

func TestRemoveExisting(t *testing.T) {
	lru := NewLRU(2, nil)
	key := "test"
	value := []byte("value")
	lru.Add(key, value)
	_, ok, _ := lru.Remove(key)
	assert.True(t, ok)
	assert.Equal(t, 0, lru.Len())
	assert.EqualValues(t, 1, lru.Metrics().Removals)
}

func TestConcurrentAccess(t *testing.T) {
	lru := NewLRU(10, nil)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			lru.Add(fmt.Sprintf("key%d", i), []byte("value"))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			lru.Get(fmt.Sprintf("key%d", i%10))
		}
	}()
	wg.Wait()
	assert.Equal(t, 10, lru.Len())
}

func TestOnEvicted(t *testing.T) {
	var evictedKey string
	onEvicted := func(key string, value []byte) bool {
		evictedKey = key
		return true
	}
	lru := NewLRU(1, onEvicted)
	lru.Add("key1", []byte("value1"))
	lru.Add("key2", []byte("value2"))
	assert.Equal(t, "key1", evictedKey)
	assert.EqualValues(t, 1, lru.Metrics().Evictions)
}

func TestClear(t *testing.T) {
	lru := NewLRU(2, nil)
	lru.Add("key1", []byte("value1"))
	lru.Add("key2", []byte("value2"))
	lru.Clear()
	assert.Equal(t, 0, lru.Len())
}

func TestMetrics(t *testing.T) {
	lru := NewLRU(2, nil)
	lru.Add("key1", []byte("value1"))
	lru.Get("key1")
	lru.Remove("key1")
	assert.EqualValues(t, 1, lru.Metrics().Inserts)
	assert.EqualValues(t, 1, lru.Metrics().Hits)
	assert.EqualValues(t, 1, lru.Metrics().Removals)
}

func TestErrorHandling(t *testing.T) {
	lru := NewLRU(1, nil)
	err := lru.Add("key", []byte("value"))
	assert.NoError(t, err)
}

func BenchmarkAdd(b *testing.B) {
	lru := NewLRU(1000, nil)
	value := []byte("value")
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		lru.Add(key, value)
	}
}

func BenchmarkGet_Hit(b *testing.B) {
	lru := NewLRU(1000, nil)
	value := []byte("value")
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		lru.Add(key, value)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		lru.Get(key)
	}
}

func BenchmarkGet_Miss(b *testing.B) {
	lru := NewLRU(1000, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		lru.Get(key)
	}
}

func BenchmarkConcurrentAdd(b *testing.B) {
	lru := NewLRU(1000, nil)
	var counter uint64
	value := []byte("value")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := atomic.AddUint64(&counter, 1)
			key := fmt.Sprintf("key%d", id)
			lru.Add(key, value)
		}
	})
}

func BenchmarkConcurrentGet(b *testing.B) {
	lru := NewLRU(1000, nil)
	value := []byte("value")
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		lru.Add(key, value)
	}
	var counter uint64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := atomic.AddUint64(&counter, 1)
			key := fmt.Sprintf("key%d", id)
			lru.Get(key)
		}
	})
}

func BenchmarkMixedOperations(b *testing.B) {
	lru := NewLRU(1000, nil)
	var counter uint64
	value := []byte("value")
	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			id := atomic.AddUint64(&counter, 1)
			if i%2 == 0 {
				key := fmt.Sprintf("key%d-%d", id, i)
				lru.Add(key, value)
			} else {
				key := fmt.Sprintf("key%d-%d", id, i-1)
				lru.Get(key)
			}
			i++
		}
	})
}
