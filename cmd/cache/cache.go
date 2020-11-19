package cache

import (

)

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte
}

type Cache struct {
	cacheMap map[string][]byte
	PushPullRequestCh chan PushPullRequest
	CacheHandlerStarted bool
}

func (cache Cache) CacheHandler(req chan PushPullRequest) {
	cache.CacheHandlerStarted = true
	for {
		select {
		case ppCacheOp := <-req:
			// pull operation
			if len(ppCacheOp.Data) <= 0 {
				ppCacheOp.ReturnPayload <- cache.cacheMap[ppCacheOp.Key]
			} else if len(ppCacheOp.Data) > 0 { // push operation
				cache.cacheMap[ppCacheOp.Key] = ppCacheOp.Data
			}
		}
	}
}

func New() Cache {
  cache := Cache{make(map[string][]byte), make(chan PushPullRequest), false}
	cache.CacheHandlerStarted = false
	go cache.CacheHandler(cache.PushPullRequestCh)
  return cache
}

func (cache Cache) AddKeyVal(key string, val []byte) {
	request := new(PushPullRequest)
	request.Key = key
	request.Data = val

  cache.PushPullRequestCh <- *request
}

func (cache Cache) GetKeyVal(key string) []byte {
	request := new(PushPullRequest)
	request.Key = key
	request.ReturnPayload = make(chan []byte)
  cache.PushPullRequestCh <- *request

	reply := false
	payload := []byte{}

	// wainting for request to be processed and retrieval of payload
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
