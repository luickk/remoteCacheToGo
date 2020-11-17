package cacheDb

import (
  "remoteCacheToGo/cmd/cache"
)

type DbOp struct {
	Operation string
	Val string
	Fullfilled bool
}

type CacheDb struct {
	Db map[string]cache.Cache
	DbOpCh chan DbOp
	CacheDbHandlerStarted bool
}

func (cacheDb CacheDb) CacheDbHandler(req chan DbOp) {
	cacheDb.CacheDbHandlerStarted = true
	var requestBuffer []DbOp

	go func(rB *[]DbOp) {
		for {
			select {
			case request := <-req:
				*rB = append(*rB, request)
			}
		}
	}(&requestBuffer)

	for {
		for _, request := range requestBuffer {
			request.Fullfilled = true
			// pull operation
			if request.Operation == "create" {
				cacheDb.Db[request.Val] = cache.New()
			} else if request.Operation == "remove" { // push operation
				delete(cacheDb.Db, request.Val)
			}
		}
	}
}

func New() CacheDb {
  cacheDb := CacheDb{make(map[string]cache.Cache), make(chan DbOp), false}
	cacheDb.CacheDbHandlerStarted = false
	go cacheDb.CacheDbHandler(cacheDb.DbOpCh)
  return cacheDb
}

func (cacheDb CacheDb) NewCache(cacheName string) {
	request := new(DbOp)
	request.Operation = "create"
	request.Val = cacheName
	request.Fullfilled = false

  cacheDb.DbOpCh <- *request
}

func (cacheDb CacheDb) RemoveCache(cacheName string) {
	request := new(DbOp)
	request.Operation = "remove"
	request.Val = cacheName
	request.Fullfilled = false

  cacheDb.DbOpCh <- *request
}

func (cacheDb CacheDb) AddEntryToCache(cacheName string, key string, val []byte) {
	cacheDb.Db[cacheName].AddKeyVal(key, val)
}

func (cacheDb CacheDb) GetEntryFromCache(cacheName string, key string) []byte {
	return cacheDb.Db[cacheName].GetKeyVal(key)
}
