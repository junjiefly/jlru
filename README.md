# jlru
kv cache with lru, support Priority, Pointer-less and zero gc. inspired by http://github.com/elastic/go-freelru

# example
```go
func main(){
	lru := NewLRU[string,[]byte](2, 100, jlru.HashXXHASH, nil)
	key := "test"
	value := []byte("value")
	lru.Add(key, value)
	lru.Get(key)
	lru.Remove(key)
}
```


# performance 

```go
cd example
go test -bench=. -benchmem
goos: linux
goarch: amd64
pkg: github.com/junjiefly/jlru/example
cpu: Intel Xeon E312xx (Sandy Bridge)
BenchmarkFreeLruAddOperation-2              	 3015058	       447.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkJLruAddOperation-2                 	 2334429	       555.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkFreeLruAddOperationWithEvict-2     	 6368730	       193.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkJLruAddOperationWithEvict-2        	 2992776	       407.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkFreeLruGetOperation-2              	 3212547	       558.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkJLruGetOperation-2                 	 2781828	       536.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkFreeLruRemoveOperation-2           	 2729623	       459.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkJLruRemoveOperation-2              	 3168922	       415.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkParallelFreeLruAddOperation-2      	 3774218	       357.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkParallelJLruAddOperation-2         	 3027030	       421.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkParallelFreeLruGetOperation-2      	 2368124	       610.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkParallelJLruGetOperation-2         	 3111256	       430.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkParallelFreeLruRemoveOperation-2   	 2155425	       543.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkParallelJLruRemoveOperation-2      	 2830405	       443.3 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/junjiefly/jlru/example	69.184s
```

