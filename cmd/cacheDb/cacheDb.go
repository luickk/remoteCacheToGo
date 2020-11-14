package cacheDb

import (
  "remoteCacheToGo/cmd/cache"
)

type CacheDb struct {
	Db map[string]cache.Cache
}

func New() CacheDb {
  cacheDb := CacheDb{make(map[string]cache.Cache)}
  return cacheDb
}

func (cacheDb CacheDb) NewCache(cacheName string) {
  cacheDb.Db[cacheName] = cache.New()
}

func (cacheDb CacheDb) RemoveCache(cacheName string) {
  delete(cacheDb.Db, cacheName)
}


func (cacheDb CacheDb) AddEntryToCache(cacheName string, key string, val []byte) {
	cacheDb.Db[cacheName].AddKeyVal(key, val)
}
