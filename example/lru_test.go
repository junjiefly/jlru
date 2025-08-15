package example

import (
	"fmt"
	"testing"
)

func BenchmarkFreeLruAddOperation(b *testing.B) {
	lru, err := newFreeKV(uint32(b.N), nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Set(keys[i], vv)
	}
}

func BenchmarkJLruAddOperation(b *testing.B) {
	lru, err := newJkv(uint32(b.N), 100, nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.SetPriority(keys[i], vv, 0)
	}
}

func BenchmarkFreeLruAddOperationWithEvict(b *testing.B) {
	lru, err := newFreeKV(10, nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Set(keys[i], vv)
	}
}

func BenchmarkJLruAddOperationWithEvict(b *testing.B) {
	lru, err := newJkv(10, 100, nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.SetPriority(keys[i], vv, 0)
	}
}

func BenchmarkFreeLruGetOperation(b *testing.B) {
	lru, err := newFreeKV(uint32(b.N), nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	for i := 0; i < b.N; i++ {
		lru.Set(keys[i], vv)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lru.Get(keys[i])
	}
}

func BenchmarkJLruGetOperation(b *testing.B) {
	lru, err := newJkv(uint32(b.N), 100, nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	for i := 0; i < b.N; i++ {
		lru.SetPriority(keys[i], vv, 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lru.Get(keys[i])
	}
}

func BenchmarkFreeLruRemoveOperation(b *testing.B) {
	lru, err := newFreeKV(uint32(b.N), nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	for i := 0; i < b.N; i++ {
		lru.Set(keys[i], vv)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lru.Delete(keys[i])
	}
}

func BenchmarkJLruRemoveOperation(b *testing.B) {
	lru, err := newJkv(uint32(b.N), 100, nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	for i := 0; i < b.N; i++ {
		lru.SetPriority(keys[i], vv, 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lru.Delete(keys[i])
	}
}

func BenchmarkParallelFreeLruAddOperation(b *testing.B) {
	lru, err := newFreeKV(uint32(b.N), nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = lru.Set(keys[i%b.N], vv)
			i++
		}
	})
}

func BenchmarkParallelJLruAddOperation(b *testing.B) {
	lru, err := newJkv(uint32(b.N), 100, nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = lru.SetPriority(keys[i%b.N], vv, 0)
			i++
		}
	})
}

func BenchmarkParallelFreeLruGetOperation(b *testing.B) {
	lru, err := newFreeKV(uint32(b.N), nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	for i := 0; i < b.N; i++ {
		lru.Set(keys[i], vv)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = lru.Get(keys[i%b.N])
			i++
		}
	})
}

func BenchmarkParallelJLruGetOperation(b *testing.B) {
	lru, err := newJkv(uint32(b.N), 100, nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	for i := 0; i < b.N; i++ {
		lru.SetPriority(keys[i], vv, 0)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = lru.Get(keys[i%b.N])
			i++
		}
	})
}

func BenchmarkParallelFreeLruRemoveOperation(b *testing.B) {
	lru, err := newFreeKV(uint32(b.N), nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	for i := 0; i < b.N; i++ {
		lru.Set(keys[i], vv)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = lru.Delete(keys[i%b.N])
			i++
		}
	})
}

func BenchmarkParallelJLruRemoveOperation(b *testing.B) {
	lru, err := newJkv(uint32(b.N), 100, nil)
	if err != nil {
		b.Fatal(err)
	}
	var vv = []byte("1234")
	var keys = make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyzkey1234567890abcdefghijklmnopqrstuvwxyz_%d", i)
	}
	for i := 0; i < b.N; i++ {
		lru.SetPriority(keys[i], vv, 0)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = lru.Delete(keys[i%b.N])
			i++
		}
	})
}
