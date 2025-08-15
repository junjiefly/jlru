package lru

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

// ------------------------------ 1. 初始化测试 ------------------------------
func TestNewPriorityLRU(t *testing.T) {
	t.Run("valid_params", func(t *testing.T) {
		// 正常初始化（容量10，最大优先级2，无驱逐回调）
		lru, err := NewPriorityLRU[string, []byte](10, 2, HashXXHASH, nil)
		assert.NoError(t, err)
		assert.NotNil(t, lru)
		assert.Equal(t, uint32(10), lru.Cap())
		assert.Equal(t, byte(2), lru.maxPriority)
	})

	t.Run("invalid_capacity", func(t *testing.T) {
		// 容量为0时返回错误
		lru, err := NewPriorityLRU[string, []byte](0, 2, HashXXHASH, nil)
		assert.Error(t, err)
		assert.Nil(t, lru)
		assert.Equal(t, errors.New("CapacityTooSmall"), err)
	})

	t.Run("min_priority_boundary", func(t *testing.T) {
		// 最小优先级为0（边界值）
		lru, err := NewPriorityLRU[string, []byte](5, 0, HashXXHASH, nil)
		assert.NoError(t, err)
		assert.Equal(t, byte(0), lru.maxPriority)
	})

	t.Run("max_priority_boundary", func(t *testing.T) {
		// 最大优先级为255（边界值）
		lru, err := NewPriorityLRU[string, []byte](5, 255, HashXXHASH, nil)
		assert.NoError(t, err)
		assert.Equal(t, byte(100), lru.maxPriority)
	})
}

// ------------------------------ 2. Add 操作测试 ------------------------------
func TestLRU_Add(t *testing.T) {
	t.Run("add_new_entry", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		err := lru.Add("key1", []byte("val1"), 0)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), lru.Metrics().Inserts)
		assert.Equal(t, uint32(1), lru.Len()) // 排除优先级标记节点
		val, ok, _ := lru.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, []byte("val1"), val)
	})

	t.Run("add_existing_key", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		lru.Add("key1", []byte("val1"), 0)
		err := lru.Add("key1", []byte("val2"), 1) // 更新值和优先级
		assert.NoError(t, err)
		assert.Equal(t, uint64(2), lru.Metrics().Inserts) // 重复添加计为Insert
		val, ok, _ := lru.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, []byte("val2"), val)
	})

	t.Run("capacity_exceed_evict", func(t *testing.T) {
		// 容量2，添加3个元素触发驱逐
		evictCount := 0
		onEvicted := func(key string, value []byte) bool {
			evictCount++
			return true
		}
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, onEvicted)
		lru.Add("key1", []byte("val1"), 0)
		lru.Add("key2", []byte("val2"), 0)
		lru.Add("key3", []byte("val3"), 0) // 触发驱逐最旧元素（key1）
		assert.Equal(t, 1, evictCount)
		assert.Equal(t, uint64(1), lru.Metrics().Evictions)
		assert.Equal(t, uint32(2), lru.Len())
		_, ok, _ := lru.Get("key1")
		assert.False(t, ok)
		_, ok, _ = lru.Get("key2")
		assert.True(t, ok)
		_, ok, _ = lru.Get("key3")
		assert.True(t, ok)

	})

	t.Run("priority_order", func(t *testing.T) {
		// 验证优先级高的元素不会被优先驱逐
		lru, _ := NewPriorityLRU[string, []byte](3, 2, HashXXHASH, nil)
		lru.Add("high", []byte("h"), 2)  // 高优先级（不易驱逐）
		lru.Add("low1", []byte("l1"), 0) // 低优先级（易易驱逐）
		lru.Add("low2", []byte("l2"), 0) // 低优先级（易驱逐）
		lru.Get("low1")                  //low1访问后，在lru队列前，low2预期被删除
		lru.Add("new", []byte("n"), 1)   // 容量满，驱逐低优先级"low"
		_, ok, _ := lru.Get("low1")
		fmt.Println("ok:", ok)
		assert.True(t, ok)
		_, ok, _ = lru.Get("low2")
		assert.False(t, ok)
		_, ok, _ = lru.Get("high")
		assert.True(t, ok)
		_, ok, _ = lru.Get("new")
		assert.True(t, ok)
	})
}

// ------------------------------ 3. Get 操作测试 ------------------------------
func TestLRU_Get(t *testing.T) {
	t.Run("get_non_existing_key", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		_, ok, err := lru.Get("key1")
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, uint64(1), lru.Metrics().Misses)
	})
	t.Run("get_existing_key", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		lru.Add("key1", []byte("val1"), 0)
		val, ok, err := lru.Get("key1")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, []byte("val1"), val)
		assert.Equal(t, uint64(1), lru.Metrics().Hits)
	})
	t.Run("get_with_priority_update", func(t *testing.T) {
		// Get 操作是否提升优先级（MoveAfter 逻辑）
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		lru.Add("key1", []byte("val1"), 0)
		lru.Add("key2", []byte("val2"), 0) // 容量满，驱逐未被Get的元素（无，因容量2）
		lru.Get("key1")                    // 提升优先级
		lru.Add("key3", []byte("val3"), 0) // 驱逐key2（因key1在lru队列前，key2在队列尾）
		_, ok, _ := lru.Get("key1")
		assert.True(t, ok)
		_, ok, _ = lru.Get("key2")
		assert.False(t, ok)
		_, ok, _ = lru.Get("key3")
		assert.True(t, ok)
	})
}

// ------------------------------ 4. Remove 操作测试 ------------------------------
func TestLRU_Remove(t *testing.T) {
	t.Run("remove_non_existing_key", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		_, ok, err := lru.Remove("key1")
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("remove_existing_key", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		lru.Add("key1", []byte("val1"), 0)
		val, ok, err := lru.Remove("key1")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, []byte("val1"), val)
		assert.Equal(t, uint64(1), lru.Metrics().Removals)
	})
	t.Run("remove_priority_mark_node", func(t *testing.T) {
		// 尝试删除优先级标记节点（节点不存在）
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		_, ok, err := lru.Remove("p0") // 标记节点key为"p0"
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("remove_oldest", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](6, 1, HashXXHASH, nil)
		lru.Add("high1", []byte("h1"), 1)
		lru.Add("high2", []byte("h2"), 1)
		lru.Add("high3", []byte("h3"), 1)
		lru.Add("high4", []byte("h4"), 1)
		lru.Add("low1", []byte("l1"), 0)
		lru.Add("low2", []byte("l2"), 0)
		lru.RemoveOldest()
		_, ok, _ := lru.Get("low1")
		assert.False(t, ok)
		_, ok, _ = lru.Get("high1")
		assert.True(t, ok)

		_, ok, _ = lru.Get("high2")
		assert.True(t, ok)
		_, ok, _ = lru.Get("high3")
		assert.True(t, ok)
		_, ok, _ = lru.Get("high4")
		assert.True(t, ok)
		_, ok, _ = lru.Get("low2")
		assert.True(t, ok)
	})
}

// ------------------------------ 5. 哈希冲突与并发测试 ------------------------------
func TestLRU_HashConflict(t *testing.T) {
	// 构造哈希冲突（通过固定hash函数返回相同值）
	hashFunc := func(s string) uint32 { return 0 } // 所有key哈希为0
	lru, _ := NewPriorityLRU[string, []byte](3, 1, hashFunc, nil)
	lru.Add("key1", []byte("val1"), 0)
	lru.Add("key2", []byte("val2"), 0)                 // 哈希冲突，存入同一bucket
	lru.Add("key3", []byte("val3"), 0)                 // 哈希冲突，存入同一bucket
	assert.Equal(t, uint64(2), lru.Metrics().Conflict) // 冲突计数+1
	val, ok, _ := lru.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, []byte("val1"), val)
	val, ok, _ = lru.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, []byte("val2"), val)
	val, ok, _ = lru.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, []byte("val3"), val)
}

func TestLRU_ConcurrentAccess(t *testing.T) {
	lru, _ := NewPriorityLRU[string, []byte](10, 2, HashXXHASH, nil)
	var wg sync.WaitGroup
	// 100个协程并发读写
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", idx)
			lru.Add(key, []byte(key), byte(idx%3))
			lru.Get(key)
			if idx%5 == 0 {
				lru.Remove(key)
			}
		}(i)
	}
	wg.Wait()
	assert.True(t, lru.Len() == 10) // 容量10，最终长度不超过容量
}

// ------------------------------ 6. 错误处理测试 ------------------------------
func TestLRU_InvalidOperations(t *testing.T) {
	t.Run("get_empty_cache", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		_, ok, err := lru.Get("key1")
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, uint64(1), lru.Metrics().Misses)
	})

	t.Run("add_invalid_priority", func(t *testing.T) {
		lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
		err := lru.Add("key1", []byte("val1"), 10) // 优先级>maxPriority(1)
		assert.NoError(t, err)                     // 内部会截断为maxPriority
		_, ok, _ := lru.Get("key1")
		assert.True(t, ok)
	})
}

// ------------------------------ 7. 指标与状态测试 ------------------------------
func TestLRU_Metrics(t *testing.T) {
	lru, _ := NewPriorityLRU[string, []byte](3, 1, HashXXHASH, nil)
	lru.Add("key1", []byte("val1"), 0) // Inserts=1
	lru.Get("key1")                    // Hits=1
	lru.Get("key2")                    // Misses=1
	lru.Add("key2", []byte("val2"), 0) // Inserts=2
	lru.Add("key3", []byte("val3"), 0) //  Inserts=3
	lru.Add("key4", []byte("val4"), 0) // 容量满，驱逐key1（Evictions=0）
	lru.Remove("key4")                 // Removals=1

	metrics := lru.Metrics()
	assert.Equal(t, uint64(4), metrics.Inserts)
	assert.Equal(t, uint64(1), metrics.Hits)
	assert.Equal(t, uint64(1), metrics.Misses)
	assert.Equal(t, uint64(1), metrics.Removals)
	assert.Equal(t, uint64(1), metrics.Evictions)
}

// ------------------------------ 8. Clear 与资源释放 ------------------------------
func TestLRU_Clear(t *testing.T) {
	lru, _ := NewPriorityLRU[string, []byte](2, 1, HashXXHASH, nil)
	lru.Add("key1", []byte("val1"), 0)
	lru.Clear()
	assert.Equal(t, uint32(0), lru.Len())
	assert.Nil(t, lru.buckets) // Clear后buckets置空
}

func BenchmarkAddOperation(b *testing.B) {
	lru, _ := NewPriorityLRU[string, []byte](1000, 5, HashXXHASH, nil)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := lru.Add("benchmark_key", []byte("benchmark_value"), 3)
			if err != nil && err.Error() != "cache is full" {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkGetOperation(b *testing.B) {
	lru, _ := NewPriorityLRU[string, []byte](1000, 5, HashXXHASH, nil)
	for i := 0; i < 1000; i++ {
		err := lru.Add(fmt.Sprintf("key%d", i), []byte("value"), 2)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lru.Get("key500")
		}
	})
}
