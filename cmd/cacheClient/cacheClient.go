package cacheClient

import (
	"net"
	"fmt"
	"strconv"
)

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte
}

type RemoteCache struct {
	conn net.Conn
	PushPullRequestCh chan *PushPullRequest
	CacheHandlerStarted bool
}

func connectToRemoteHandler(address string, port int) (bool, net.Conn) {
  c, err := net.Dial("tcp", address+":"+strconv.Itoa(port))
  if err != nil {
    fmt.Println(err)
    return false, c
  }
  return true, c
}

func (cache RemoteCache) pushPullRequestHandler() {
  cache.CacheHandlerStarted = true
	for {
		select {
		case ppCacheOp := <-cache.PushPullRequestCh:
			if len(ppCacheOp.Data) <= 0 { // pull operation
				cache.conn.Write(append([]byte(ppCacheOp.Key + "->-"), []byte("\r")...))
				ppCacheOp.ReturnPayload <- []byte("lol")
			} else if len(ppCacheOp.Data) > 0 { // push operation
				cache.conn.Write(append(append([]byte(ppCacheOp.Key + "-<-"), ppCacheOp.Data...), []byte("\r")...))
			}
		}
	}
}

func New(address string, port int) RemoteCache {
	connected, conn := connectToRemoteHandler(address, port)
  if !connected {
    fmt.Println("could not connect to given address")
  }
  cache := RemoteCache{conn, make(chan *PushPullRequest), false}
  go cache.pushPullRequestHandler()
	return cache
}

func (cache RemoteCache) AddKeyVal(key string, val []byte) {
	request := new(PushPullRequest)
	request.Key = key
	request.Data = val

  cache.PushPullRequestCh <- request

	request = nil
}

func (cache RemoteCache) GetKeyVal(key string) []byte {
	request := new(PushPullRequest)
	request.Key = key
	request.ReturnPayload = make(chan []byte)
  cache.PushPullRequestCh <- request

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

	request = nil
	return payload
}