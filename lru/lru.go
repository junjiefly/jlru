package lru

import (
	"errors"
	"fmt"
	"github.com/cespare/xxhash/v2"
	jlist "github.com/junjiefly/jlru/list"
	"math"
	"sync"
	"sync/atomic"
)

const emptyBucket = math.MaxUint32
const invalidIdx = math.MaxUint32
const maxEntryPriority = 100

type ListMetrics struct {
	Inserts   uint64
	Evictions uint64
	Removals  uint64
	Hits      uint64
	Misses    uint64
	Conflict  uint64
	Errors    uint64
}

func HashXXHASH(s string) uint32 {
	return uint32(xxhash.Sum64String(s))
}

// HashKeyCallback is the function that creates a hash from the passed key.
type HashKeyCallback[K comparable] func(K) uint32

type OnEvictCallback[K comparable, V any] func(K, V) bool

// LRU  a lru supports priority.
type LRU[K comparable, V any] struct {
	metrics ListMetrics
	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted OnEvictCallback[K, V]

	ll          *jlist.List[K, V]
	buckets     []uint32
	cap         uint32
	pos         []uint32
	maxPriority byte
	sync.RWMutex
	hashFunc HashKeyCallback[K]
}

func NewPriorityLRU[K comparable, V any](capacity int, maxPriority byte, hashFunc HashKeyCallback[K], onEvicted OnEvictCallback[K, V]) (*LRU[K, V], error) {
	if capacity == 0 {
		return nil, errors.New("CapacityTooSmall")
	}
	if maxPriority > maxEntryPriority {
		maxPriority = maxEntryPriority
	}
	lru := &LRU[K, V]{
		OnEvicted:   onEvicted,
		cap:         uint32(capacity),
		buckets:     make([]uint32, capacity),
		pos:         make([]uint32, maxPriority+2),
		maxPriority: maxPriority,
		hashFunc:    hashFunc,
	}
	lru.ll = jlist.NewList[K, V](capacity + int(maxPriority) + 2)
	for pos := range lru.pos {
		e, err := lru.ll.PushFront(*new(K), *new(V), byte(pos))
		if err != nil {
			return nil, err
		}
		e.Flag = 1
		lru.pos[pos] = e.Idx()
	}
	for k := range lru.buckets {
		lru.buckets[k] = emptyBucket
	}
	return lru, nil
}

func (lru *LRU[K, V]) getBucketPos(hashId uint32) uint32 {
	return hashId % uint32(len(lru.buckets))
}

func (lru *LRU[K, V]) getEntryInBuk(pos uint32, key K) (*jlist.Entry[K, V], bool, error) {
	if pos >= lru.cap {
		return nil, false, errors.New("getEntryInBuk err: InvalidPos")
	}

	//nextIdx := lru.buckets[pos]
	//seq := 0
	//for nextIdx != emptyBucket {
	//	entry, err := lru.ll.Entry(nextIdx) //A <A ->A
	//	if err != nil {
	//		atomic.AddUint64(&lru.metrics.Errors, 1)
	//		return nil, false, fmt.Errorf("addEntryInBuk err: %s", err.Error())
	//	}
	//	nextIdx = entry.ConflictNext
	//	if nextIdx == lru.buckets[pos] {
	//		break
	//	}
	//	seq++
	//}

	startIdx := lru.buckets[pos]
	if startIdx == emptyBucket {
		return nil, false, nil
	}
	idx := startIdx
	for idx != emptyBucket {
		e, err := lru.ll.Entry(idx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return nil, false, fmt.Errorf("getEntryInBuk err: %s", err.Error())
		}
		if e.Key == key {
			return e, true, nil
		}
		idx = e.ConflictNext
		if idx == startIdx {
			break
		}
	}
	return nil, false, nil
}

func (lru *LRU[K, V]) addEntryInBuk(pos uint32, newIdx uint32) error {
	if pos >= lru.cap {
		return errors.New("addEntryInBuk err: InvalidPos")
	}
	newEntry, err := lru.ll.Entry(newIdx)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return fmt.Errorf("addEntryInBuk err: %s", err.Error())
	}
	startIdx := lru.buckets[pos]
	if startIdx == emptyBucket {
		lru.buckets[pos] = newIdx
		newEntry.ConflictPrev = newIdx
		newEntry.ConflictNext = newIdx
	} else {
		if startIdx == newIdx {
			return nil
		}
		headEntry, err := lru.ll.Entry(startIdx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("addEntryInBuk err: %s", err.Error())
		}
		tailIdx := headEntry.ConflictPrev
		tailEntry, err := lru.ll.Entry(tailIdx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("addEntryInBuk err: %s", err.Error())
		}
		tailEntry.ConflictNext = newIdx
		newEntry.ConflictPrev = tailIdx
		newEntry.ConflictNext = startIdx
		headEntry.ConflictPrev = newIdx
		atomic.AddUint64(&lru.metrics.Conflict, 1)
	}
	//
	//nextIdx := lru.buckets[pos]
	//seq := 0
	//for nextIdx != emptyBucket {
	//	entry, err := lru.ll.Entry(nextIdx)
	//	if err != nil {
	//		atomic.AddUint64(&lru.metrics.Errors, 1)
	//		return fmt.Errorf("addEntryInBuk err: %s", err.Error())
	//	}
	//	nextIdx = entry.ConflictNext
	//	if nextIdx == lru.buckets[pos] {
	//		break
	//	}
	//	seq++
	//}
	return nil
}

func (lru *LRU[K, V]) removeEntryFromBuk(pos uint32, delIdx uint32) error {
	if pos >= lru.cap {
		return errors.New("removeEntryFromBuk err: invalidPos")
	}
	delEntry, err := lru.ll.Entry(delIdx)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return fmt.Errorf("removeEntryFromBuk err: %s", err.Error())
	}
	startIdx := lru.buckets[pos]
	if startIdx == invalidIdx {
		return nil
	}
	headEntry, err := lru.ll.Entry(startIdx)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return fmt.Errorf("removeEntryFromBuk err: %s", err.Error())
	}
	tailIdx := headEntry.ConflictPrev
	if delIdx == startIdx && delIdx == tailIdx {
		lru.buckets[pos] = emptyBucket
		delEntry.ConflictNext = invalidIdx
		delEntry.ConflictPrev = invalidIdx
		return nil
	}
	tailEntry, err := lru.ll.Entry(tailIdx)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return fmt.Errorf("removeEntryFromBuk err: %s", err.Error())
	}
	if delIdx == startIdx {
		nextIdx := delEntry.ConflictNext
		nextEntry, err := lru.ll.Entry(nextIdx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("removeEntryFromBuk err: %s", err.Error())
		}
		lru.buckets[pos] = nextIdx
		nextEntry.ConflictPrev = tailIdx
		tailEntry.ConflictNext = nextIdx
	} else if delIdx == tailIdx {
		prevIdx := delEntry.ConflictPrev
		prevEntry, err := lru.ll.Entry(prevIdx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("removeEntryFromBuk err: %s", err.Error())
		}
		prevEntry.ConflictNext = startIdx
		headEntry.ConflictPrev = prevIdx
	} else {
		prevIdx := delEntry.ConflictPrev
		prevEntry, err := lru.ll.Entry(prevIdx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("removeEntryFromBuk err: %s", err.Error())
		}
		nextIdx := delEntry.ConflictNext
		nextEntry, err := lru.ll.Entry(nextIdx)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("removeEntryFromBuk err: %s", err.Error())
		}
		prevEntry.ConflictNext = nextIdx
		nextEntry.ConflictPrev = prevIdx
	}
	delEntry.ConflictNext = invalidIdx
	delEntry.ConflictPrev = invalidIdx

	//nextIdx := lru.buckets[pos]
	//seq := 0
	//for nextIdx != emptyBucket {
	//	entry, err := lru.ll.Entry(nextIdx) //A <A ->A
	//	if err != nil {
	//		atomic.AddUint64(&lru.metrics.Errors, 1)
	//		return fmt.Errorf("removeEntryFromBuk err: %s", err.Error())
	//	}
	//	//fmt.Println("TT list in removeEntryFromBuk  pos:", pos, "seq:", seq, "idx:", entry.Idx(), "vs:", nextIdx, "key:", entry.Key)
	//	nextIdx = entry.ConflictNext
	//	if nextIdx == lru.buckets[pos] {
	//		break
	//	}
	//	seq++
	//}

	return nil
}

func (lru *LRU[K, V]) hashToPos(key K) (hashId uint32, bukPos uint32) {
	hashId = lru.hashFunc(key)
	return hashId, lru.getBucketPos(hashId)
}

// Add adds a value to the cache.
func (lru *LRU[K, V]) Add(key K, value V, priority byte) error {
	if priority > lru.maxPriority {
		priority = lru.maxPriority
	}
	hashId, bukPos := lru.hashToPos(key)
	lru.Lock()
	defer lru.Unlock()
	e, ok, err := lru.getEntryInBuk(bukPos, key)
	if err != nil {
		return fmt.Errorf("add err: %s", err.Error())
	}
	markNode, err := lru.getPriorityMarkNode(priority + 1)
	if err != nil {
		return fmt.Errorf("add err: %s", err.Error())
	}
	if ok {
		err = lru.ll.MoveAfter(e, markNode)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("add err: %s", err.Error())
		}
		e.Priority = priority
		e.Key = key
		e.HashId = hashId
		e.Value = value
		err = lru.ll.UpdateEntry(e.Idx(), e)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("add err: %s", err.Error())
		}
		atomic.AddUint64(&lru.metrics.Inserts, 1)
		return nil
	}
	if lru.ll.Len() >= lru.ll.Cap() {
		lru.removeOldest()
	}
	ele, err := lru.ll.InsertAfter(key, value, markNode)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return fmt.Errorf("add err: %s", err.Error())
	}
	ele.HashId = hashId
	err = lru.addEntryInBuk(bukPos, ele.Idx())
	if err != nil {
		lru.ll.Remove(ele)
		return fmt.Errorf("add err: %s", err.Error())
	}
	atomic.AddUint64(&lru.metrics.Inserts, 1)
	return nil
}

func (lru *LRU[K, V]) AddToBack(key K, value V, priority byte) error {
	if priority > lru.maxPriority {
		priority = lru.maxPriority
	}
	hashId, bukPos := lru.hashToPos(key)
	lru.Lock()
	defer lru.Unlock()
	e, ok, err := lru.getEntryInBuk(bukPos, key)
	if err != nil {
		return fmt.Errorf("addToBack err: %s", err.Error())
	}
	markNode, err := lru.getPriorityMarkNode(priority)
	if err != nil {
		return fmt.Errorf("addToBack err: %s", err.Error())
	}
	if ok {
		err = lru.ll.MoveBefore(e, markNode)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("addToBack err: %s", err.Error())
		}
		e.HashId = hashId
		e.Priority = priority
		e.Key = key
		e.Value = value
		err = lru.ll.UpdateEntry(e.Idx(), e)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return fmt.Errorf("addToBack err: %s", err.Error())
		}
		atomic.AddUint64(&lru.metrics.Inserts, 1)
		return nil
	}
	if lru.ll.Len() >= lru.ll.Cap() {
		lru.removeOldest()
	}
	ele, err := lru.ll.InsertBefore(key, value, markNode)
	if err != nil {
		return fmt.Errorf("addToBack err: %s", err.Error())
	}
	ele.HashId = hashId
	err = lru.addEntryInBuk(bukPos, ele.Idx())
	if err != nil {
		lru.ll.Remove(ele)
		return fmt.Errorf("addToBack err: %s", err.Error())
	}
	atomic.AddUint64(&lru.metrics.Inserts, 1)
	return nil
}

func (lru *LRU[K, V]) getPriorityMarkNode(priority byte) (*jlist.Entry[K, V], error) {
	posIdx := lru.pos[priority]
	markNode, err := lru.ll.Entry(posIdx)
	if err != nil {
		return nil, fmt.Errorf("getPriorityMarkNode err: %s", err.Error())
	}
	return markNode, nil
}

// Get looks up a key's value from the cache.
func (lru *LRU[K, V]) Get(key K) (value V, ok bool, err error) {
	_, bukPos := lru.hashToPos(key)
	lru.Lock()
	defer lru.Unlock()
	e, ok, err := lru.getEntryInBuk(bukPos, key)
	if err != nil {
		return value, false, fmt.Errorf("get err: %s", err.Error())
	}
	if ok {
		markNode, err := lru.getPriorityMarkNode(e.Priority + 1)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return value, false, fmt.Errorf("get err: %s", err.Error())
		}
		atomic.AddUint64(&lru.metrics.Hits, 1)
		value = e.Value
		err = lru.ll.MoveAfter(e, markNode)
		if err != nil {
			atomic.AddUint64(&lru.metrics.Errors, 1)
			return value, true, fmt.Errorf("get err: %s", err.Error())
		}
		return value, true, nil
	}
	atomic.AddUint64(&lru.metrics.Misses, 1)
	return value, false, nil
}

// Has looks up a key's value from the cache.
func (lru *LRU[K, V]) Has(key K) (value V, ok bool, err error) {
	_, bukPos := lru.hashToPos(key)
	lru.RLock()
	defer lru.RUnlock()
	ele, ok, err := lru.getEntryInBuk(bukPos, key)
	if err != nil {
		return value, false, fmt.Errorf("has err: %s", err.Error())
	}
	if ok {
		atomic.AddUint64(&lru.metrics.Conflict, 1)
		return ele.Value, false, nil
	}
	return value, false, nil
}

// Remove removes the provided key from the cache.
func (lru *LRU[K, V]) Remove(key K) (value V, ok bool, err error) {
	_, bukPos := lru.hashToPos(key)
	lru.Lock()
	defer lru.Unlock()
	e, ok, err := lru.getEntryInBuk(bukPos, key)
	if err != nil {
		return value, false, fmt.Errorf("remove err: %s", err.Error())
	}
	if !ok {
		return value, ok, nil
	}
	if e.Flag > 0 {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return value, false, errors.New("remove err: not user node")
	}
	if e.Key != key {
		atomic.AddUint64(&lru.metrics.Conflict, 1)
		return value, false, errors.New("remove err: key conflict")
	}
	value = e.Value
	err = lru.removeElement(e, false)
	if err != nil {
		atomic.AddUint64(&lru.metrics.Errors, 1)
		return value, false, fmt.Errorf("remove err: %s", err.Error())
	}
	atomic.AddUint64(&lru.metrics.Removals, 1)
	return value, true, nil
}

func (lru *LRU[K, V]) removeOldest() {
	ele := lru.oldest()
	if ele != nil {
		if lru.removeElement(ele, true) == nil {
			atomic.AddUint64(&lru.metrics.Evictions, 1)
		}
	}
	return
}

// RemoveOldest removes the oldest item from the cache.
func (lru *LRU[K, V]) RemoveOldest() bool {
	lru.Lock()
	defer lru.Unlock()
	ele := lru.oldest()
	if ele != nil {
		if lru.removeElement(ele, true) == nil {
			atomic.AddUint64(&lru.metrics.Removals, 1)
			return true
		}
		return false
	}
	return false
}

func (lru *LRU[K, V]) oldest() *jlist.Entry[K, V] {
	var i byte
	for i = 0; i < lru.maxPriority; i++ {
		markNode, err := lru.getPriorityMarkNode(i)
		if err != nil {
			return nil
		}
		e, err := lru.ll.Entry(markNode.Prev())
		if err != nil {
			return nil
		}
		if e.Flag == 0 {
			return e
		}
		continue
	}
	return nil
}

func (lru *LRU[K, V]) removeElement(e *jlist.Entry[K, V], evict bool) error {
	if e == nil {
		return nil
	}
	if e.Flag > 0 {
		return errors.New("removeElement err: not user node")
	}
	bukPos := lru.getBucketPos(e.HashId)
	if evict && lru.OnEvicted != nil {
		if lru.OnEvicted(e.Key, e.Value) {
			err := lru.removeEntryFromBuk(bukPos, e.Idx())
			if err != nil {
				return fmt.Errorf("removeElement err:%s", err.Error())
			}
			_, err = lru.ll.Remove(e)
			if err != nil {
				return fmt.Errorf("removeElement err:%s", err.Error())
			}
		}
		return nil
	}
	err := lru.removeEntryFromBuk(bukPos, e.Idx())
	if err != nil {
		return fmt.Errorf("removeElement err:%s", err.Error())
	}
	_, err = lru.ll.Remove(e)
	if err != nil {
		return fmt.Errorf("removeElement err:%s", err.Error())
	}
	return nil
}

// Len returns the number of items in the cache.
func (lru *LRU[K, V]) Len() uint32 {
	lru.RLock()
	defer lru.RUnlock()
	if lru.ll == nil {
		return 0
	}
	return lru.ll.Len() - uint32(lru.maxPriority) - 2
}

func (lru *LRU[K, V]) Metrics() ListMetrics {
	lru.RLock()
	defer lru.RUnlock()
	return lru.metrics
}

func (lru *LRU[K, V]) Iterate() (keys []K, values []V, priority []byte) {
	lru.RLock()
	defer lru.RUnlock()
	for pos, startIdx := range lru.buckets {
		if startIdx == emptyBucket {
			continue
		}
		idx := startIdx
		for lru.buckets[pos] != emptyBucket {
			e, err := lru.ll.Entry(idx)
			if err == nil {
				idx = e.ConflictNext
			} else {
				break
			}
			if idx == startIdx {
				break
			}
		}
	}
	return lru.ll.Iterate()
}

func (lru *LRU[K, V]) Cap() uint32 {
	return lru.ll.Cap() - uint32(lru.maxPriority) - 2
}

// Clear purges all stored items from the cache.
func (lru *LRU[K, V]) Clear() {
	for pos, startIdx := range lru.buckets {
		if startIdx == emptyBucket {
			continue
		}
		idx := startIdx
		for lru.buckets[pos] != emptyBucket {
			e, err := lru.ll.Entry(idx)
			if err == nil {
				if lru.OnEvicted != nil {
					if !lru.OnEvicted(e.Key, e.Value) {
						continue
					}
				}
				idx = e.ConflictNext
				_ = lru.removeEntryFromBuk(uint32(pos), e.Idx())
				_, _ = lru.ll.Remove(e)
			}
		}
	}
	lru.ll.Clear()
	lru.ll = nil
	lru.buckets = nil
}
