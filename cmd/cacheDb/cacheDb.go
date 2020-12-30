package cacheDb

import (
  "remoteCacheToGo/cmd/cache"
  "remoteCacheToGo/pkg/util"
)

type DbOp struct {
	Operation string
	Val string
}

type CacheDb struct {
	Db map[string]cache.Cache
	DbOpCh chan *DbOp
}

func (cacheDb CacheDb) CacheDbHandler(cacheDbOp chan *DbOp) {
	for {
		select {
		case dbOp := <-cacheDbOp:
			if dbOp.Operation == "create" {
				cacheDb.Db[dbOp.Val] = cache.New()
			} else if dbOp.Operation == "remove" { // push operation
				delete(cacheDb.Db, dbOp.Val)
			}
  	}
  }
}

func New() CacheDb {
  cacheDb := CacheDb{ make(map[string]cache.Cache), make(chan *DbOp) }
	go cacheDb.CacheDbHandler(cacheDb.DbOpCh)
  return cacheDb
}

func (cacheDb CacheDb) NewCache(cacheName string) bool {
  if util.CharacterWhiteList(cacheName) {
  	request := new(DbOp)
  	request.Operation = "create"
  	request.Val = cacheName

    cacheDb.DbOpCh <- request

  	request = nil
    return true
  }
  return false
}

func (cacheDb CacheDb) RemoveCache(cacheName string) {
  if util.CharacterWhiteList(cacheName) {
  	request := new(DbOp)
  	request.Operation = "remove"
  	request.Val = cacheName

    cacheDb.DbOpCh <- request

  	request = nil
  }
}
