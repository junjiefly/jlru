package list

import (
	"reflect"
	"testing"
)

func TestNewList(t *testing.T) {
	var capacity = 5
	list := NewList[string, int](capacity)
	if list == nil {
		t.Fatal("Failed to create new list")
	}
	if list.cap != uint32(capacity) {
		t.Errorf("Expected data length %d, got %d", capacity, list.cap)
	}
	if len(list.freeIdx) != capacity {
		t.Errorf("Expected freeIdx size %d, got %d", capacity, len(list.freeIdx))
	}
}

func TestPushFront(t *testing.T) {
	list := NewList[string, int](3)
	testCases := []struct {
		key   string
		value int
	}{
		{"A", 1},
		{"B", 2},
		{"C", 3},
	}

	for i, tc := range testCases {
		entry, err := list.PushFront(tc.key, tc.value, 0)
		if err != nil {
			t.Fatalf("Test %d: %v", i, err)
		}
		if entry.Key != tc.key || entry.Value != tc.value {
			t.Errorf("Test %d: Invalid entry values", i)
		}
	}

	if list.Len() != 3 {
		t.Errorf("Expected length 3, got %d", list.Len())
	}
}

func TestRemove(t *testing.T) {
	list := NewList[string, int](3)
	e1, _ := list.PushFront("A", 1, 0)
	e2, _ := list.PushFront("B", 2, 0)
	e3, _ := list.PushFront("C", 3, 0)

	// Normal removal
	val, err := list.Remove(e2)
	if err != nil {
		t.Fatal(err)
	}
	if val != 2 {
		t.Errorf("Expected value 2, got %d", val)
	}
	if list.Len() != 2 {
		t.Errorf("Expected length 2, got %d", list.Len())
	}

	// Remove head
	val, err = list.Remove(e3)
	if val != 3 {
		t.Errorf("Expected value 3, got %d", val)
	}
	if list.Len() != 1 {
		t.Errorf("Expected length 1, got %d", list.Len())
	}

	// Remove last node
	val, err = list.Remove(e1)
	if val != 1 {
		t.Errorf("Expected value 1, got %d", val)
	}
	if list.Len() != 0 {
		t.Errorf("Expected length 0, got %d", list.Len())
	}
}

func TestMoveOperations(t *testing.T) {
	list := NewList[string, int](3)
	e1, _ := list.PushBack("A", 1, 0)
	e2, _ := list.PushBack("B", 2, 0)
	_, _ = list.PushBack("C", 3, 0)

	list.MoveToFront(e2)
	expected := []int{2, 1, 3}
	_, values, _ := list.Iterate()

	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Unexpected order after MoveToFront")
	}

	list.MoveToBack(e1)
	expected = []int{2, 3, 1}
	_, values, _ = list.Iterate()

	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Unexpected order after MoveToBack")
	}
}

func TestInsertOperations(t *testing.T) {
	list := NewList[string, int](5)
	e1, _ := list.PushBack("A", 1, 0)
	e2, _ := list.PushBack("B", 2, 0)

	// Insert before
	_, err := list.InsertBefore("C", 0, e1)
	if err != nil {
		t.Fatal(err)
	}
	expected := []int{0, 1, 2}
	_, values, _ := list.Iterate()
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Unexpected order after InsertBefore")
	}

	// Insert after
	_, err = list.InsertAfter("D", 3, e2)
	if err != nil {
		t.Fatal(err)
	}
	expected = []int{0, 1, 2, 3}

	_, values, _ = list.Iterate()
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Unexpected order after InsertAfter")
	}
}

func TestErrorConditions(t *testing.T) {
	list := NewList[string, int](2)
	e1, _ := list.PushFront("A", 1, 0)

	// Memory exhaustion
	_, err := list.PushFront("B", 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = list.PushFront("C", 3, 0)
	if err == nil {
		t.Error("Expected memory pool exhausted error")
	}

	// Invalid node operations
	_, err = list.Remove(e1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = list.Remove(e1)
	if err == nil {
		t.Error("Expected invalid node error")
	}
}

func TestIterate(t *testing.T) {
	list := NewList[string, int](4)
	e1, _ := list.PushBack("A", 1, 0)
	e2, _ := list.PushBack("B", 2, 0)
	list.MoveToFront(e1)
	list.MoveToBack(e2)

	_, values, _ := list.Iterate()

	expected := []int{1, 2}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Unexpected iteration result")
	}
}

func Test0Iterate(t *testing.T) {
	list := NewList[string, int](1)
	e, _ := list.PushBack("A", 1, 0)
	list.Remove(e)
	vs, _, _ := list.Iterate()
	for range vs {
		t.Errorf("Expected len 0")
	}
}

func TestUpdateEntry(t *testing.T) {
	list := NewList[string, int](3)
	e1, _ := list.PushFront("A", 1, 0)
	entry, _ := list.Entry(e1.idx)
	entry.Value = 100

	err := list.UpdateEntry(e1.idx, entry)
	if err != nil {
		t.Fatal(err)
	}

	updatedEntry, _ := list.Entry(e1.idx)
	if updatedEntry.Value != 100 {
		t.Errorf("Expected value 100, got %d", updatedEntry.Value)
	}
}

func TestClear(t *testing.T) {
	list := NewList[string, int](3)
	list.PushFront("A", 1, 0)
	list.Clear()
	if list.Len() != 0 || list.size != 0 || len(list.freeIdx) != 0 {
		t.Error("Clear operation failed")
	}
}

func BenchmarkNewList(b *testing.B) {
	capacity := 100
	for i := 0; i < b.N; i++ {
		NewList[int, int](capacity)
	}
}

func BenchmarkPushFront(b *testing.B) {
	l := NewList[int, int](100)
	for i := 0; i < b.N; i++ {
		l.PushFront(i, i, 1)
	}
}

func BenchmarkPushBack(b *testing.B) {
	l := NewList[int, int](100)
	for i := 0; i < b.N; i++ {
		l.PushBack(i, i, 1)
	}
}

func BenchmarkRemove(b *testing.B) {
	l := NewList[int, int](100)
	for i := 0; i < 10; i++ {
		l.PushFront(i, i, 1)
	}
	for i := 0; i < b.N; i++ {
		front := l.Front()
		l.Remove(front)
	}
}

func BenchmarkIterate(b *testing.B) {
	l := NewList[int, int](100)
	for i := 0; i < 10; i++ {
		l.PushFront(i, i, 1)
	}
	for i := 0; i < b.N; i++ {
		l.Iterate()
	}
}

// more test

func TestInsertHead(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []int // 插入的数据源
		expected []int // 预期的链表顺序（头到尾）
	}{
		{
			name:     "空链表插入头部1",
			inputs:   []int{1},
			expected: []int{1},
		},
		{
			name:     "多元素插入头部2-0",
			inputs:   []int{1, 2},
			expected: []int{2, 1},
		},
		{
			name:     "多元素插入头部（逆序）",
			inputs:   []int{1, 2, 3},
			expected: []int{3, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, data := range tt.inputs {
				list.PushFront(data, data, byte(data))
			}
			result, _, _ := list.Iterate()
			// 验证长度
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			// 验证顺序
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func TestInsertTail(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []int // 插入的数据源
		expected []int // 预期的链表顺序（头到尾）
	}{
		{
			name:     "空链表插入尾部",
			inputs:   []int{1},
			expected: []int{1},
		},
		{
			name:     "多元素插入尾部（顺序）",
			inputs:   []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, data := range tt.inputs {
				list.PushBack(data, data, byte(data))
			}
			result, _, _ := list.Iterate()
			// 验证长度
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			// 验证顺序
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func TestInsertBefore(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int // 初始链表（尾插）
		target   int   // 目标元素
		data     int   // 插入的数据
		expected []int // 预期结果（头到尾）
		wantErr  bool  // 是否期望错误
	}{
		{
			name:     "插入到节点前面1",
			initial:  []int{1},
			target:   1,
			data:     2,
			expected: []int{2, 1},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面2-0",
			initial:  []int{1, 2},
			target:   1,
			data:     3,
			expected: []int{3, 1, 2},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面2-1",
			initial:  []int{1, 2},
			target:   2,
			data:     3,
			expected: []int{1, 3, 2},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面3-0",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     4,
			expected: []int{4, 1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面3-1",
			initial:  []int{1, 2, 3},
			target:   2,
			data:     4,
			expected: []int{1, 4, 2, 3},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面3-2",
			initial:  []int{1, 2, 3},
			target:   3,
			data:     4,
			expected: []int{1, 2, 4, 3},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面4-0",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     5,
			expected: []int{5, 1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面4-1",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     5,
			expected: []int{1, 5, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面4-2",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     5,
			expected: []int{1, 2, 5, 3, 4},
			wantErr:  false,
		},
		{
			name:     "插入到节点前面4-3",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     5,
			expected: []int{1, 2, 3, 5, 4},
			wantErr:  false,
		},
		{
			name:     "目标元素不存在",
			initial:  []int{1, 2, 3},
			target:   4,
			data:     5,
			expected: []int{1, 2, 3},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, d := range tt.initial {
				list.PushBack(d, d, byte(d))
			}
			target, _ := list.Find(tt.target)
			node, err := list.InsertBefore(tt.data, tt.data, target)
			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("预期错误%v，实际错误%v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			// 验证节点非空
			if node == nil {
				t.Fatal("预期节点非空，实际为空")
			}
			// 验证节点数据
			if node.Key != tt.data {
				t.Errorf("预期节点数据%v，实际数据%v", tt.data, node.Key)
			}
			// 验证链表顺序
			result, _, _ := list.Iterate()
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func TestInsertAfter(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int // 初始链表（尾插）
		target   int   // 目标元素
		data     int   // 插入的数据
		expected []int // 预期结果（头到尾）
		wantErr  bool  // 是否期望错误
	}{
		{
			name:     "插入到节点后面1",
			initial:  []int{1},
			target:   1,
			data:     2,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "插入到节点后面2-0",
			initial:  []int{1, 2},
			target:   1,
			data:     3,
			expected: []int{1, 3, 2},
			wantErr:  false,
		},
		{
			name:     "插入到节点后面2-1",
			initial:  []int{1, 2},
			target:   2,
			data:     3,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "插入到节点后面3-0",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     4,
			expected: []int{1, 4, 2, 3},
			wantErr:  false,
		},
		{
			name:     "插入到节点后面3-1",
			initial:  []int{1, 2, 3},
			target:   2,
			data:     4,
			expected: []int{1, 2, 4, 3},
			wantErr:  false,
		},
		{
			name:     "插入到节点后面3-2",
			initial:  []int{1, 2, 3},
			target:   3,
			data:     4,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "目标元素不存在",
			initial:  []int{1, 2, 3},
			target:   4,
			data:     5,
			expected: []int{1, 2, 3},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, d := range tt.initial {
				list.PushBack(d, d, byte(d))
			}
			target, _ := list.Find(tt.target)
			node, err := list.InsertAfter(tt.data, tt.data, target)
			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("预期错误%v，实际错误%v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			// 验证节点数据
			if node.Key != tt.data {
				t.Errorf("预期节点数据%v，实际数据%v", tt.data, node.Key)
			}
			// 验证链表顺序
			result, _, _ := list.Iterate()
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func TestMoveBefore(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int // 初始链表（尾插）
		target   int   // 目标元素
		data     int   // 插入的数据
		expected []int // 预期结果（头到尾）
		wantErr  bool  // 是否期望错误
	}{
		{
			name:     "移动节点前面1",
			initial:  []int{1},
			target:   1,
			data:     1,
			expected: []int{1},
			wantErr:  false,
		},
		{
			name:     "移动节点前面2-0",
			initial:  []int{1, 2},
			target:   1,
			data:     1,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点前面2-1",
			initial:  []int{1, 2},
			target:   1,
			data:     2,
			expected: []int{2, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点前面2-2",
			initial:  []int{1, 2},
			target:   2,
			data:     1,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点前面2-3",
			initial:  []int{1, 2},
			target:   2,
			data:     2,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点前面3-0",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     1,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点前面3-1",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     2,
			expected: []int{2, 1, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点前面3-2",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     3,
			expected: []int{3, 1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点前面3-3",
			initial:  []int{1, 2, 3},
			target:   3,
			data:     1,
			expected: []int{2, 1, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点前面3-4",
			initial:  []int{1, 2, 3},
			target:   3,
			data:     2,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点前面3-5",
			initial:  []int{1, 2, 3},
			target:   3,
			data:     3,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-0",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     1,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-1",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     2,
			expected: []int{2, 1, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-2",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     3,
			expected: []int{3, 1, 2, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-3",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     4,
			expected: []int{4, 1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-4",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     1,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-5",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     2,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-6",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     3,
			expected: []int{1, 3, 2, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-7",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     4,
			expected: []int{1, 4, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-8",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     1,
			expected: []int{2, 1, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-9",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     2,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-10",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     3,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-11",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     4,
			expected: []int{1, 2, 4, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-12",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     1,
			expected: []int{2, 3, 1, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-13",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     2,
			expected: []int{1, 3, 2, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-14",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     3,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面4-15",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     4,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-0",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     1,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-1",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     2,
			expected: []int{2, 1, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-2",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     3,
			expected: []int{3, 1, 2, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-3",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     4,
			expected: []int{4, 1, 2, 3, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-4",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     5,
			expected: []int{5, 1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-5",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     1,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-6",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     2,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-7",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     3,
			expected: []int{1, 3, 2, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-8",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     4,
			expected: []int{1, 4, 2, 3, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-9",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     5,
			expected: []int{1, 5, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-10",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     1,
			expected: []int{2, 1, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-11",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     2,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-12",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     3,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-13",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     4,
			expected: []int{1, 2, 4, 3, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-14",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     5,
			expected: []int{1, 2, 5, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-15",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     1,
			expected: []int{2, 3, 1, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-16",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     2,
			expected: []int{1, 3, 2, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-17",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     3,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-18",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     4,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-19",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     5,
			expected: []int{1, 2, 3, 5, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-20",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     1,
			expected: []int{2, 3, 4, 1, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-21",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     2,
			expected: []int{1, 3, 4, 2, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-22",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     3,
			expected: []int{1, 2, 4, 3, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-23",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     4,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面5-24",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     5,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-0",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     1,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-1",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     2,
			expected: []int{2, 1, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-2",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     3,
			expected: []int{3, 1, 2, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-3",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     4,
			expected: []int{4, 1, 2, 3, 5, 6},
			wantErr:  false,
		},

		{
			name:     "移动节点前面6-4",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     5,
			expected: []int{5, 1, 2, 3, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-5",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     6,
			expected: []int{6, 1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-6",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     1,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-7",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     2,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},

		{
			name:     "移动节点前面6-8",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     3,
			expected: []int{1, 3, 2, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-9",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     4,
			expected: []int{1, 4, 2, 3, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-10",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     5,
			expected: []int{1, 5, 2, 3, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-11",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     6,
			expected: []int{1, 6, 2, 3, 4, 5},
			wantErr:  false,
		},

		{
			name:     "移动节点前面6-12",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     1,
			expected: []int{2, 1, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-13",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     2,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-14",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     3,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-15",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     4,
			expected: []int{1, 2, 4, 3, 5, 6},
			wantErr:  false,
		},

		{
			name:     "移动节点前面6-16",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     5,
			expected: []int{1, 2, 5, 3, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-17",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     6,
			expected: []int{1, 2, 6, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-18",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     1,
			expected: []int{2, 3, 1, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-19",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     2,
			expected: []int{1, 3, 2, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-20",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     3,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-21",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     4,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-22",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     5,
			expected: []int{1, 2, 3, 5, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-23",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     6,
			expected: []int{1, 2, 3, 6, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-24",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     1,
			expected: []int{2, 3, 4, 1, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-25",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     2,
			expected: []int{1, 3, 4, 2, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-26",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     3,
			expected: []int{1, 2, 4, 3, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-27",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     4,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-28",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     5,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-29",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     6,
			expected: []int{1, 2, 3, 4, 6, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-30",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     1,
			expected: []int{2, 3, 4, 5, 1, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-31",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     2,
			expected: []int{1, 3, 4, 5, 2, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-32",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     3,
			expected: []int{1, 2, 4, 5, 3, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-33",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     4,
			expected: []int{1, 2, 3, 5, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-34",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     5,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面6-35",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     6,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点前面12",
			initial:  []int{1, 2, 3},
			target:   4,
			data:     2,
			expected: []int{2, 1, 3},
			wantErr:  true,
		},
		{
			name:     "移动节点前面13",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     4,
			expected: []int{2, 1, 3},
			wantErr:  true,
		},
		{
			name:     "移动节点前面14",
			initial:  []int{1, 2, 3},
			target:   4,
			data:     5,
			expected: []int{2, 1, 3},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, d := range tt.initial {
				list.PushBack(d, d, byte(d))
			}
			target, _ := list.Find(tt.target)
			node, _ := list.Find(tt.data)
			err := list.MoveBefore(node, target)
			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("预期错误%v，实际错误%v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			// 验证节点数据
			if node.Key != tt.data {
				t.Errorf("预期节点数据%v，实际数据%v", tt.data, node.Key)
			}
			// 验证链表顺序
			result, _, _ := list.Iterate()
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func TestMoveAfter(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int // 初始链表（尾插）
		target   int   // 目标元素
		data     int   // 插入的数据
		expected []int // 预期结果（头到尾）
		wantErr  bool  // 是否期望错误
	}{
		{
			name:     "移动节点后面1",
			initial:  []int{1},
			target:   1,
			data:     1,
			expected: []int{1},
			wantErr:  false,
		},
		{
			name:     "移动节点后面2-0",
			initial:  []int{1, 2},
			target:   1,
			data:     1,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点后面2-1",
			initial:  []int{1, 2},
			target:   1,
			data:     2,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点后面2-2",
			initial:  []int{1, 2},
			target:   2,
			data:     1,
			expected: []int{2, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点后面2-3",
			initial:  []int{1, 2},
			target:   2,
			data:     2,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点后面3-0",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     1,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点后面3-1",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     2,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点后面3-2",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     3,
			expected: []int{1, 3, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点后面3-3",
			initial:  []int{1, 2, 3},
			target:   3,
			data:     1,
			expected: []int{2, 3, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点后面3-4",
			initial:  []int{1, 2, 3},
			target:   3,
			data:     2,
			expected: []int{1, 3, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点后面3-5",
			initial:  []int{1, 2, 3},
			target:   3,
			data:     3,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-0",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     1,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-1",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     2,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-2",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     3,
			expected: []int{1, 3, 2, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-3",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			data:     4,
			expected: []int{1, 4, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-4",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     1,
			expected: []int{2, 1, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-5",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     2,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-6",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     3,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-7",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			data:     4,
			expected: []int{1, 2, 4, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-8",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     1,
			expected: []int{2, 3, 1, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-9",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     2,
			expected: []int{1, 3, 2, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-10",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     3,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-11",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			data:     4,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-12",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     1,
			expected: []int{2, 3, 4, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-13",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     2,
			expected: []int{1, 3, 4, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-14",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     3,
			expected: []int{1, 2, 4, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点后面4-15",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			data:     4,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-0",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     1,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-1",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     2,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-2",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     3,
			expected: []int{1, 3, 2, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-3",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     4,
			expected: []int{1, 4, 2, 3, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-4",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			data:     5,
			expected: []int{1, 5, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-5",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     1,
			expected: []int{2, 1, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-6",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     2,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-7",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     3,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-8",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     4,
			expected: []int{1, 2, 4, 3, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-9",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			data:     5,
			expected: []int{1, 2, 5, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-10",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     1,
			expected: []int{2, 3, 1, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-11",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     2,
			expected: []int{1, 3, 2, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-12",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     3,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-13",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     4,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-14",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			data:     5,
			expected: []int{1, 2, 3, 5, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-15",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     1,
			expected: []int{2, 3, 4, 1, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-16",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     2,
			expected: []int{1, 3, 4, 2, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-17",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     3,
			expected: []int{1, 2, 4, 3, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-18",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     4,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-19",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			data:     5,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-20",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     1,
			expected: []int{2, 3, 4, 5, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-21",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     2,
			expected: []int{1, 3, 4, 5, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-22",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     3,
			expected: []int{1, 2, 4, 5, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-23",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     4,
			expected: []int{1, 2, 3, 5, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面5-24",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			data:     5,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-0",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     1,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-1",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     2,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-2",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     3,
			expected: []int{1, 3, 2, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-3",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     4,
			expected: []int{1, 4, 2, 3, 5, 6},
			wantErr:  false,
		},

		{
			name:     "移动节点后面6-4",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     5,
			expected: []int{1, 5, 2, 3, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-5",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   1,
			data:     6,
			expected: []int{1, 6, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-6",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     1,
			expected: []int{2, 1, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-7",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     2,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},

		{
			name:     "移动节点后面6-8",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     3,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-9",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     4,
			expected: []int{1, 2, 4, 3, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-10",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     5,
			expected: []int{1, 2, 5, 3, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-11",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   2,
			data:     6,
			expected: []int{1, 2, 6, 3, 4, 5},
			wantErr:  false,
		},

		{
			name:     "移动节点后面6-12",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     1,
			expected: []int{2, 3, 1, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-13",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     2,
			expected: []int{1, 3, 2, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-14",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     3,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-15",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     4,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},

		{
			name:     "移动节点后面6-16",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     5,
			expected: []int{1, 2, 3, 5, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-17",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   3,
			data:     6,
			expected: []int{1, 2, 3, 6, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-18",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     1,
			expected: []int{2, 3, 4, 1, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-19",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     2,
			expected: []int{1, 3, 4, 2, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-20",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     3,
			expected: []int{1, 2, 4, 3, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-21",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     4,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-22",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     5,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-23",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   4,
			data:     6,
			expected: []int{1, 2, 3, 4, 6, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-24",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     1,
			expected: []int{2, 3, 4, 5, 1, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-25",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     2,
			expected: []int{1, 3, 4, 5, 2, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-26",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     3,
			expected: []int{1, 2, 4, 5, 3, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-27",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     4,
			expected: []int{1, 2, 3, 5, 4, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-28",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     5,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-29",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   5,
			data:     6,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-30",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     1,
			expected: []int{2, 3, 4, 5, 6, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-31",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     2,
			expected: []int{1, 3, 4, 5, 6, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-32",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     3,
			expected: []int{1, 2, 4, 5, 6, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-33",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     4,
			expected: []int{1, 2, 3, 5, 6, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-34",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     5,
			expected: []int{1, 2, 3, 4, 6, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点后面6-35",
			initial:  []int{1, 2, 3, 4, 5, 6},
			target:   6,
			data:     6,
			expected: []int{1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "移动节点后面12",
			initial:  []int{1, 2, 3},
			target:   4,
			data:     2,
			expected: []int{2, 1, 3},
			wantErr:  true,
		},
		{
			name:     "移动节点后面13",
			initial:  []int{1, 2, 3},
			target:   1,
			data:     4,
			expected: []int{2, 1, 3},
			wantErr:  true,
		},
		{
			name:     "移动节点后面14",
			initial:  []int{1, 2, 3},
			target:   4,
			data:     5,
			expected: []int{2, 1, 3},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, d := range tt.initial {
				list.PushBack(d, d, byte(d))
			}
			target, _ := list.Find(tt.target)
			node, _ := list.Find(tt.data)
			err := list.MoveAfter(node, target)
			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("预期错误%v，实际错误%v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			// 验证节点数据
			if node.Key != tt.data {
				t.Errorf("预期节点数据%v，实际数据%v", tt.data, node.Key)
			}
			// 验证链表顺序
			result, _, _ := list.Iterate()
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int // 初始链表（尾插）
		target   int   // 要删除的元素
		expected []int // 预期结果（头到尾）
		wantErr  bool  // 是否期望错误
	}{
		{
			name:     "删除中间节点",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			expected: []int{1, 3, 4},
			wantErr:  false,
		},
		{
			name:     "删除头节点",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			expected: []int{2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "删除尾节点",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "删除唯一节点",
			initial:  []int{1},
			target:   1,
			expected: []int{},
			wantErr:  false,
		},
		{
			name:     "目标元素不存在",
			initial:  []int{1, 2, 3},
			target:   4,
			expected: []int{1, 2, 3},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, d := range tt.initial {
				list.PushBack(d, d, byte(d))
			}
			target, _ := list.Find(tt.target)
			_, err := list.remove(target)
			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("预期错误%v，实际错误%v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			// 验证链表顺序
			result, _, _ := list.Iterate()
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func TestMoveToHead(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int // 初始链表（尾插）
		target   int   // 要移动的元素
		expected []int // 预期结果（头到尾）
		wantErr  bool  // 是否期望错误
	}{
		{
			name:     "移动节点到头部1",
			initial:  []int{1},
			target:   1,
			expected: []int{1},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部2-0",
			initial:  []int{1, 2},
			target:   1,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部2-1",
			initial:  []int{1, 2},
			target:   2,
			expected: []int{2, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部3-0",
			initial:  []int{1, 2, 3},
			target:   1,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部3-1",
			initial:  []int{1, 2, 3},
			target:   2,
			expected: []int{2, 1, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部3-2",
			initial:  []int{1, 2, 3},
			target:   3,
			expected: []int{3, 1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部4-0",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部4-1",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			expected: []int{2, 1, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部4-2",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			expected: []int{3, 1, 2, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部4-3",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			expected: []int{4, 1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部5-0",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部5-1",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			expected: []int{2, 1, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部5-2",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			expected: []int{3, 1, 2, 4, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部5-3",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			expected: []int{4, 1, 2, 3, 5},
			wantErr:  false,
		},
		{
			name:     "移动节点到头部5-4",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			expected: []int{5, 1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "节点不存在",
			initial:  []int{1, 2, 3, 4},
			target:   5,
			expected: []int{1, 2, 3, 4},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, d := range tt.initial {
				list.PushBack(d, d, byte(d))
			}
			target, _ := list.Find(tt.target)
			err := list.MoveToFront(target)

			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("预期错误%v，实际错误%v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			// 验证链表顺序
			result, _, _ := list.Iterate()
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段（移动不改变长度）
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func TestMoveToTail(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int // 初始链表（尾插）
		target   int   // 要移动的元素
		expected []int // 预期结果（头到尾）
		wantErr  bool  // 是否期望错误
	}{
		{
			name:     "移动节点到尾部1",
			initial:  []int{1},
			target:   1,
			expected: []int{1},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部2-0",
			initial:  []int{1, 2},
			target:   1,
			expected: []int{2, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部2-1",
			initial:  []int{1, 2},
			target:   2,
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部3-0",
			initial:  []int{1, 2, 3},
			target:   1,
			expected: []int{2, 3, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部3-1",
			initial:  []int{1, 2, 3},
			target:   2,
			expected: []int{1, 3, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部3-2",
			initial:  []int{1, 2, 3},
			target:   3,
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部4-0",
			initial:  []int{1, 2, 3, 4},
			target:   1,
			expected: []int{2, 3, 4, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部4-1",
			initial:  []int{1, 2, 3, 4},
			target:   2,
			expected: []int{1, 3, 4, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部4-2",
			initial:  []int{1, 2, 3, 4},
			target:   3,
			expected: []int{1, 2, 4, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部4-3",
			initial:  []int{1, 2, 3, 4},
			target:   4,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部5-0",
			initial:  []int{1, 2, 3, 4, 5},
			target:   1,
			expected: []int{2, 3, 4, 5, 1},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部5-1",
			initial:  []int{1, 2, 3, 4, 5},
			target:   2,
			expected: []int{1, 3, 4, 5, 2},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部5-2",
			initial:  []int{1, 2, 3, 4, 5},
			target:   3,
			expected: []int{1, 2, 4, 5, 3},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部5-3",
			initial:  []int{1, 2, 3, 4, 5},
			target:   4,
			expected: []int{1, 2, 3, 5, 4},
			wantErr:  false,
		},
		{
			name:     "移动节点到尾部5-4",
			initial:  []int{1, 2, 3, 4, 5},
			target:   5,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "节点不存在",
			initial:  []int{1, 2, 3, 4},
			target:   5,
			expected: []int{1, 2, 3, 4},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := 10
			list := NewList[int, int](capacity)
			for _, d := range tt.initial {
				list.PushBack(d, d, byte(d))
			}
			target, ok := list.Find(tt.target)
			if !ok {
				if tt.wantErr == true {
					return
				}
				t.Fatalf("没有找到指定元素:%v", tt.target)
			}
			err := list.MoveToBack(target)
			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("预期错误%v，实际错误%v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			// 验证链表顺序
			result, _, _ := list.Iterate()
			if len(result) != len(tt.expected) {
				t.Fatalf("预期长度%d，实际长度%d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引%d预期值%v，实际值%v", i, tt.expected[i], result[i])
				}
			}
			// 验证Size字段（移动不改变长度）
			if list.size != uint32(len(tt.expected)) {
				t.Errorf("预期Size%d，实际Size%d", len(tt.expected), list.size)
			}
		})
	}
}

func BenchmarkInsertHead(b *testing.B) {
	// 1. 构建千万级元素的链表（尾插，模拟已有大链表）
	capacity := 10000000
	list := NewList[int, int](b.N + capacity)
	for i := 0; i < capacity; i++ {
		list.PushBack(i, i, byte(i))
	}
	// 2. 重置计时器（不计算构建链表的时间）
	b.ResetTimer()
	// 3. 测试插入头部的性能（O(1)）
	for i := capacity; i < capacity+b.N; i++ {
		list.PushFront(i, i, byte(i))
	}
}

func BenchmarkInsertTail(b *testing.B) {
	// 1. 构建千万级元素的链表（尾插）
	capacity := 10000000
	list := NewList[int, int](b.N + capacity)
	for i := 0; i < capacity; i++ {
		list.PushBack(i, i, byte(i))
	}
	// 2. 重置计时器
	b.ResetTimer()
	// 3. 测试插入尾部的性能（O(1)）
	for i := capacity; i < capacity+b.N; i++ {
		list.PushBack(i, i, byte(i))
	}
}
func BenchmarkInsertBeforeMiddle(b *testing.B) {
	// 1. 构建千万级元素的链表（尾插）
	capacity := 10000000
	list := NewList[uint32, uint32](b.N + capacity)
	for i := 0; i < capacity; i++ {
		list.PushBack(uint32(i), uint32(i), byte(i))
	}
	// 2. 找到中间节点（第5e6个元素，模拟插入到中间）
	middle, ok := list.Find(list.size / 2)
	if !ok {
		b.Fatalf("中间节点未找到")
	}
	// 3. 重置计时器
	b.ResetTimer()
	// 4. 测试插入到中间节点前面的性能（O(1)插入，但O(n)查找已包含在构建阶段）
	for i := (capacity); i < (capacity + b.N); i++ {
		_, err := list.InsertBefore(uint32(i), uint32(i), middle)
		if err != nil {
			b.Fatal(err)
		}
	}
}
