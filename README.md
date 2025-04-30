# jlru
kv cache with lru, Pointer-less and zero gc

# example
```go
func main(){
	lru := NewLRU(2, nil)
	key := "test"
	value := []byte("value")
	lru.Add(key, value)
  lru.Get(key)
  lru.Remove(key)
}
```
