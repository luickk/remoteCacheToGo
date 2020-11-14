package cache

type Cache struct {
	cache map[string][]byte
}

func New() Cache {
  cache := Cache{make(map[string][]byte)}
  return cache
}

func (cache Cache) AddKeyVal(key string, val []byte) {
  cache.cache[key] = val
}
