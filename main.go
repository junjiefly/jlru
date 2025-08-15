package main

import (
	"fmt"
	jlru "github.com/junjiefly/jlru/lru"
	"sync/atomic"
	"time"
)

func main() {
	var fileCount = 10
	var capacity = fileCount - 1
	var data = []byte("12345")
	lru, err := jlru.NewPriorityLRU[string, []byte](capacity, 100, jlru.HashXXHASH, nil)
	if err != nil {
		fmt.Println("init priority lru err:", err)
		return
	}
	var files = make([]string, fileCount)
	for ii := 0; ii < fileCount; ii++ {
		files[ii] = fmt.Sprintf("file_%d", ii)
	}
	var now = time.Now()
	for ii := 0; ii < fileCount; ii++ {
		_ = lru.Add(files[ii], data, byte(fileCount-ii))
	}
	fmt.Println("priority lru set cost:", time.Since(now), "cnt:", fileCount, "avg", time.Since(now).Nanoseconds()/int64(fileCount))
	var getOk int64
	now = time.Now()
	for ii := 0; ii < fileCount; ii++ {
		_, ok, err := lru.Get(files[ii])
		if err != nil {
			fmt.Println("Get file:", files[ii], "err:", err)
			continue
		}
		if ok {
			atomic.AddInt64(&getOk, 1)
		}
	}
	fmt.Println("priority lru get cost:", time.Since(now), "cnt:", fileCount, "getOk:", getOk, "Conflict:", lru.Metrics().Conflict, "avg", time.Since(now).Nanoseconds()/int64(fileCount))
}
