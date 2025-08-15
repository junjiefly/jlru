package list

import (
	"errors"
	"math"
)

const invalidPos = math.MaxUint32

type Entry[K comparable, V any] struct {
	Flag     byte   //类型标记位,非0时表示是一个被标记的节点[flag != 0 means this is not a user node]
	Priority byte   //优先级，地优先级意味着更容易被驱逐[priority for the entry，low priority means easier to be evicted]
	prev     uint32 //当前节点在LRU双向链表的前一个节点[prev node in the lru double linked list]
	next     uint32 //当前节点在LRU双向链表的下一个节点[next node in the lru double linked list]
	idx      uint32 //block序号
	HashId   uint32 //哈希值
	Key      K      //键
	Value    V      //值

	ConflictPrev uint32 //当前节点在冲突双向链表的前一个节点[prev node in the conflict double linked list]
	ConflictNext uint32 //当前节点在冲突双向链表的下一个节点[next node in the conflict double linked list]
}

func (e Entry[K, V]) Match(key K) bool {
	return e.Key == key
}

func (e Entry[K, V]) Idx() uint32 {
	return e.idx
}

func (e Entry[K, V]) Prev() uint32 {
	return e.prev
}

type List[K comparable, V any] struct {
	cap     uint32        //容量
	data    []Entry[K, V] //包含所有block的切片
	freeIdx []uint32      //空闲block序号
	head    uint32        //lru队列头,
	tail    uint32        //lru队列尾,靠近尾部的节点更容易被驱逐[node close to the tail means easier to be evicted]
	size    uint32        //当前lru队列长度
}

func NewList[K comparable, V any](capacity int) *List[K, V] {
	ll := &List[K, V]{
		freeIdx: make([]uint32, capacity),
		head:    invalidPos,
		tail:    invalidPos,

		size: 0,
		cap:  uint32(capacity),
	}
	ll.data = make([]Entry[K, V], capacity)
	for i := 0; i < capacity; i++ {
		ll.freeIdx[i] = uint32(i)
	}
	return ll
}

func (l *List[K, V]) getNodeIdx() (uint32, bool) {
	if len(l.freeIdx) == 0 {
		return invalidPos, false
	}
	idx := l.freeIdx[len(l.freeIdx)-1]       // get last
	l.freeIdx = l.freeIdx[:len(l.freeIdx)-1] // eject last
	l.data[idx].ConflictPrev = invalidPos
	l.data[idx].ConflictNext = invalidPos
	l.data[idx].prev = invalidPos
	l.data[idx].next = invalidPos
	return idx, true
}

func (l *List[K, V]) putNodeIdx(idx uint32) {
	l.freeIdx = append(l.freeIdx, idx) // append to end
}

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *List[K, V]) Len() uint32 {
	return l.size
}

func (l *List[K, V]) Cap() uint32 {
	return l.cap
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List[K, V]) Front() *Entry[K, V] {
	if l.size == 0 {
		return nil
	}
	if l.head != invalidPos {
		return &l.data[l.head]
	}
	return nil
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[K, V]) Back() *Entry[K, V] {
	if l.size == 0 {
		return nil
	}
	if l.tail != invalidPos {
		return &l.data[l.tail]
	}
	return nil
}

func (l *List[K, V]) remove(e *Entry[K, V]) (V, error) {
	if e == nil {
		var empty V
		return empty, errors.New("node null")
	}
	if e.next == invalidPos || e.prev == invalidPos {
		return e.Value, errors.New("unknown node")
	}
	if l.cap <= e.idx || e.idx < 0 {
		return e.Value, errors.New("invalid node")
	}
	node := l.data[e.idx]
	if node.prev == invalidPos || node.next == invalidPos {
		return e.Value, errors.New("invalid node")
	}
	if e.next != node.next || e.prev != node.prev {
		return e.Value, errors.New("list changed")
	}
	value := e.Value
	l.data[node.prev].next = node.next
	l.data[node.next].prev = node.prev

	if l.head == e.idx {
		l.head = node.next
	}
	if l.tail == e.idx {
		l.tail = node.prev
	}
	l.data[e.idx].prev = invalidPos
	l.data[e.idx].next = invalidPos
	l.data[e.idx].ConflictPrev = invalidPos
	l.data[e.idx].ConflictNext = invalidPos
	l.putNodeIdx(e.idx)
	l.size--
	if l.size == 0 {
		l.head = invalidPos
		l.tail = invalidPos
	}
	return value, nil
}

func (l *List[K, V]) Remove(e *Entry[K, V]) (V, error) {
	return l.remove(e)
}

func (l *List[K, V]) PushFront(key K, value V, priority byte) (*Entry[K, V], error) {
	idx, ok := l.getNodeIdx()
	if !ok {
		return nil, errors.New("memory pool exhausted")
	}
	l.data[idx].Priority = priority
	l.data[idx].Key = key
	l.data[idx].Value = value
	l.data[idx].idx = idx

	if l.head != invalidPos && l.tail != invalidPos {
		l.data[l.head].prev = idx
		l.data[l.tail].next = idx
		l.data[idx].next = l.head
		l.data[idx].prev = l.tail
		l.head = idx
	} else {
		l.data[idx].next = idx
		l.data[idx].prev = idx
		l.head = idx
		l.tail = idx
	}
	l.size++
	return &l.data[idx], nil
}

func (l *List[K, V]) PushBack(key K, value V, priority byte) (*Entry[K, V], error) {
	idx, ok := l.getNodeIdx()
	if !ok {
		return nil, errors.New("memory pool exhausted")
	}
	l.data[idx].Priority = priority
	l.data[idx].Key = key
	l.data[idx].Value = value
	l.data[idx].idx = idx

	if l.head != invalidPos && l.tail != invalidPos {
		l.data[l.tail].next = idx
		l.data[l.head].prev = idx
		l.data[idx].prev = l.tail
		l.data[idx].next = l.head
		l.tail = idx
	} else {
		l.data[idx].next = idx
		l.data[idx].prev = idx
		l.head = idx
		l.tail = idx
	}
	l.size++
	return &l.data[idx], nil
}

func (l *List[K, V]) InsertBefore(k K, v V, mark *Entry[K, V]) (*Entry[K, V], error) {
	return l.insertBefore(k, v, mark)
}

func (l *List[K, V]) insertBefore(k K, v V, mark *Entry[K, V]) (*Entry[K, V], error) {
	idx, ok := l.getNodeIdx()
	if !ok {
		return nil, errors.New("memory pool exhausted")
	}
	if mark == nil {
		return nil, errors.New("invalid mark node")
	}
	if l.cap <= mark.idx || mark.idx < 0 {
		return nil, errors.New("invalid mark node")
	}
	markNode := l.data[mark.idx]
	if markNode.prev == invalidPos || markNode.next == invalidPos {
		return nil, errors.New("invalid node")
	}
	l.data[idx].Priority = mark.Priority
	l.data[idx].Key = k
	l.data[idx].Value = v
	l.data[idx].idx = idx

	l.data[idx].next = markNode.idx
	l.data[idx].prev = markNode.prev

	l.data[markNode.prev].next = idx
	l.data[markNode.idx].prev = idx

	if l.head == markNode.idx || l.head == invalidPos {
		l.head = idx
	}
	l.size++
	return &l.data[idx], nil
}

func (l *List[K, V]) InsertAfter(k K, v V, mark *Entry[K, V]) (*Entry[K, V], error) {
	return l.insertAfter(k, v, mark)
}

func (l *List[K, V]) insertAfter(k K, v V, mark *Entry[K, V]) (*Entry[K, V], error) {
	idx, ok := l.getNodeIdx()
	if !ok {
		return nil, errors.New("memory pool exhausted")
	}
	if mark == nil {
		return nil, errors.New("invalid mark node")
	}
	if l.cap <= mark.idx || mark.idx < 0 {
		return nil, errors.New("invalid mark node")
	}
	markNode := l.data[mark.idx]
	if markNode.prev == invalidPos || markNode.next == invalidPos {
		return nil, errors.New("invalid node")
	}
	l.data[idx].Priority = mark.Priority - 1
	l.data[idx].Key = k
	l.data[idx].Value = v
	l.data[idx].idx = idx

	l.data[idx].prev = markNode.idx
	l.data[idx].next = markNode.next

	l.data[markNode.next].prev = idx
	l.data[markNode.idx].next = idx

	if l.tail == markNode.idx || l.tail == invalidPos {
		l.tail = idx
	}
	l.size++
	return &l.data[idx], nil
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[K, V]) MoveToFront(e *Entry[K, V]) error {
	if e == nil {
		return errors.New("invalid node")
	}
	if l.cap <= e.idx || e.idx < 0 {
		return errors.New("invalid node")
	}
	markNode := l.data[e.idx]
	if markNode.prev == invalidPos || markNode.next == invalidPos {
		return errors.New("invalid node")
	}
	if e.idx == l.head {
		return nil
	}
	if e.idx == l.tail {
		l.head = e.idx
		l.tail = e.prev
	} else {
		l.data[e.prev].next = e.next
		l.data[e.next].prev = e.prev

		l.data[l.tail].next = e.idx
		l.data[l.head].prev = e.idx

		l.data[e.idx].next = l.head
		l.data[e.idx].prev = l.tail

		l.head = e.idx
	}
	return nil
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[K, V]) MoveToBack(e *Entry[K, V]) error {
	if e == nil {
		return errors.New("invalid node")
	}
	if l.cap <= e.idx || e.idx < 0 {
		return errors.New("invalid node")
	}
	markNode := l.data[e.idx]
	if markNode.prev == invalidPos || markNode.next == invalidPos {
		return errors.New("invalid node")
	}
	if e.idx == l.tail {
		return nil
	}
	if e.idx == l.head {
		l.head = e.next
		l.tail = e.idx
	} else {
		l.data[e.prev].next = e.next
		l.data[e.next].prev = e.prev

		l.data[l.tail].next = e.idx
		l.data[l.head].prev = e.idx

		l.data[e.idx].next = l.head
		l.data[e.idx].prev = l.tail

		l.tail = e.idx
	}
	return nil
}

// MoveAfter moves element e after the mark.
func (l *List[K, V]) MoveAfter(e *Entry[K, V], mark *Entry[K, V]) error {
	if e == nil {
		return errors.New("invalid node")
	}
	if l.cap <= e.idx || e.idx < 0 {
		return errors.New("invalid node")
	}
	if mark == nil {
		return errors.New("invalid mark node")
	}
	if l.cap <= mark.idx || mark.idx < 0 {
		return errors.New("invalid mark node")
	}
	node := l.data[e.idx]
	if node.prev == invalidPos || node.next == invalidPos {
		return errors.New("invalid node")
	}
	markNode := l.data[mark.idx]
	if markNode.prev == invalidPos || markNode.next == invalidPos {
		return errors.New("invalid mark node")
	}
	if e.idx == markNode.idx {
		return nil
	}
	if e.prev == markNode.idx {
		if e.idx == l.head && l.tail == markNode.idx {
		} else {
			return nil
		}
	}
	if e.idx == l.head && mark.idx == l.tail {
		l.head = e.next
		l.tail = e.idx
	} else {
		mNext := markNode.next
		ePrev := node.prev
		eNext := node.next
		if e.idx == l.tail {
			l.tail = ePrev
		} else if markNode.idx == l.tail {
			l.tail = e.idx
		}
		if e.idx == l.head {
			l.head = node.next
		}

		l.data[e.idx].prev = markNode.idx
		l.data[e.idx].next = mNext

		l.data[eNext].prev = ePrev
		l.data[ePrev].next = eNext

		l.data[mNext].prev = e.idx
		l.data[markNode.idx].next = e.idx
	}
	return nil
}

// MoveBefore moves element e before the mark.
func (l *List[K, V]) MoveBefore(e *Entry[K, V], mark *Entry[K, V]) error {
	if e == nil {
		return errors.New("invalid node")
	}
	if l.cap <= e.idx || e.idx < 0 {
		return errors.New("invalid node")
	}
	if mark == nil {
		return errors.New("invalid mark node")
	}
	if l.cap <= mark.idx || mark.idx < 0 {
		return errors.New("invalid mark node")
	}
	node := l.data[e.idx]
	if node.prev == invalidPos || node.next == invalidPos {
		return errors.New("invalid node")
	}
	markNode := l.data[mark.idx]
	if markNode.prev == invalidPos || markNode.next == invalidPos {
		return errors.New("invalid mark node")
	}
	if e.idx == markNode.idx {
		return nil
	}
	if e.next == markNode.idx {
		if e.idx == l.tail && l.head == markNode.idx {
		} else {
			return nil
		}
	}
	if e.idx == l.tail && mark.idx == l.head {
		l.head = e.idx
		l.tail = e.prev
	} else {
		mPrev := markNode.prev
		ePrev := node.prev
		eNext := node.next
		if e.idx == l.tail {
			l.tail = ePrev
		}
		if e.idx == l.head {
			l.head = eNext
		} else if markNode.idx == l.head {
			l.head = e.idx
		}
		l.data[e.idx].prev = mPrev
		l.data[e.idx].next = markNode.idx

		l.data[ePrev].next = eNext
		l.data[eNext].prev = ePrev

		l.data[mPrev].next = e.idx
		l.data[markNode.idx].prev = e.idx
	}
	return nil
}

func (l *List[K, V]) Iterate() (keys []K, vals []V, prioritys []byte) {
	current := l.head
	for current != invalidPos {
		if l.data[current].Flag == 0 {
			keys = append(keys, l.data[current].Key)
			vals = append(vals, l.data[current].Value)
			prioritys = append(prioritys, l.data[current].Priority)
		}
		current = l.data[current].next
		if current == l.head {
			break
		}
	}
	return keys, vals, prioritys
}

func (l *List[K, V]) Entry(idx uint32) (*Entry[K, V], error) {
	if l.cap <= idx || idx < 0 || idx == invalidPos {
		return nil, errors.New("invalid node")
	}
	markNode := l.data[idx]
	if markNode.prev == invalidPos || markNode.next == invalidPos {
		return nil, errors.New("invalid node")
	}
	return &l.data[idx], nil
}

func (l *List[K, V]) UpdateEntry(idx uint32, e *Entry[K, V]) error {
	if l.cap <= idx || idx < 0 || idx == invalidPos {
		return errors.New("invalid node")
	}
	markNode := l.data[idx]
	if markNode.prev == invalidPos || markNode.next == invalidPos {
		return errors.New("invalid node")
	}
	if e == nil {
		return errors.New("invalid node")
	}
	l.data[idx].Priority = e.Priority
	l.data[idx].Key = e.Key
	l.data[idx].HashId = e.HashId
	l.data[idx].Value = e.Value
	return nil
}

func (l *List[K, V]) Clear() {
	l.data = nil
	l.freeIdx = nil
	l.head = invalidPos
	l.tail = invalidPos
	l.size = 0
}

func equal[K comparable](a, b K) bool {
	return a == b
}

func (l *List[K, V]) Find(k K) (*Entry[K, V], bool) {
	curr := l.head
	for curr != invalidPos {
		if equal(l.data[curr].Key, k) {
			return &l.data[curr], true
		}
		curr = l.data[curr].next
		if curr == l.head {
			return nil, false
		}
	}
	return nil, false
}
