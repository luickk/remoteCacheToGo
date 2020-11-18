package cacheDb

import (
  "remoteCacheToGo/cmd/cache"
)

type DbOp struct {
	Operation string
	Val string
}

type CachePushPullRequest struct {
	Operation string
	CacheName string

	Key string
	ReturnPayload chan []byte
	Data []byte
}

type CacheDb struct {
	Db map[string]cache.Cache
	DbOpCh chan DbOp
	CacheOpCh chan CachePushPullRequest
	CacheDbHandlerStarted bool
}

func (cacheDb CacheDb) CacheDbHandler(cacheDbOp chan DbOp,  cacheOp chan CachePushPullRequest) {
	cacheDb.CacheDbHandlerStarted = true

	for {
		select {
		case dbOp := <-cacheDbOp:
			if dbOp.Operation == "create" {
				cacheDb.Db[dbOp.Val] = cache.New()
			} else if dbOp.Operation == "remove" { // push operation
				delete(cacheDb.Db, dbOp.Val)
			}
		case cOp := <-cacheOp:
			if cOp.Operation == "add" {
				cacheDb.Db[cOp.CacheName].AddKeyVal(cOp.Key, cOp.Data)
			} else if cOp.Operation == "get" {
				cOp.ReturnPayload <- cacheDb.Db[cOp.CacheName].GetKeyVal(cOp.Key)
			}
		}
	}
}

func New() CacheDb {
  cacheDb := CacheDb{make(map[string]cache.Cache), make(chan DbOp), make(chan CachePushPullRequest), false}
	cacheDb.CacheDbHandlerStarted = false
	go cacheDb.CacheDbHandler(cacheDb.DbOpCh, cacheDb.CacheOpCh)
  return cacheDb
}

func (cacheDb CacheDb) NewCache(cacheName string) {
	request := new(DbOp)
	request.Operation = "create"
	request.Val = cacheName

  cacheDb.DbOpCh <- *request
}

func (cacheDb CacheDb) RemoveCache(cacheName string) {
	request := new(DbOp)
	request.Operation = "remove"
	request.Val = cacheName

  cacheDb.DbOpCh <- *request
}

func (cacheDb CacheDb) AddEntryToCache(cacheName string, key string, val []byte) {
	request := new(CachePushPullRequest)
	request.Operation = "add"
	request.CacheName = cacheName
	request.Key = key
	request.Data = val

  cacheDb.CacheOpCh <- *request
}

func (cacheDb CacheDb) GetEntryFromCache(cacheName string, key string) []byte {
	request := new(CachePushPullRequest)
	request.Operation = "get"
	request.CacheName = cacheName
	request.Key = key

	cacheDb.CacheOpCh <- *request

	reply := false
	payload := []byte{}

	// waiting for request to be processed and retrieval of payload
	for !reply {
		select {
		case liveDataRes := <-request.ReturnPayload:
			payload = liveDataRes
			reply = true
			break
		}
	}
	return payload
}
