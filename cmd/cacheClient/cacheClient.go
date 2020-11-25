package cacheClient

import (
	"net"
	"fmt"
	"strconv"
	"bytes"
	"bufio"
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

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 1024

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
	data := make([]byte, tcpConnBuffer)
	cacheListener := make(chan *PushPullRequest)
	var ppCacheOpBuffer []*PushPullRequest

	go func(conn net.Conn, cacheListener chan *PushPullRequest) {
		for {
			n, err := bufio.NewReader(conn).Read(data)
			if err != nil {
				fmt.Println(err)
			}
			data = data[:n]
			netDataSeperated := bytes.Split(data, []byte("\r"))
			if err != nil {
				fmt.Println(err)
			}

			for _, data := range netDataSeperated {
				if len(data) >= 1 {
						dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 3)
						key := string(dataDelimSplitByte[0])
						operation := string(dataDelimSplitByte[1])
						payload := dataDelimSplitByte[2]
						if operation == ">" {
							request := new(PushPullRequest)
							request.Key = key
							request.Data = payload
							cacheListener <- request
							request = nil
						}
				}
			}
		}
	}(cache.conn, cacheListener)

	for {
		select {
		case ppCacheOp := <-cache.PushPullRequestCh:
			ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
			if len(ppCacheOp.Data) <= 0 { // pull operation
				cache.conn.Write(append([]byte(ppCacheOp.Key + "->-"), []byte("\r")...))
			} else if len(ppCacheOp.Data) > 0 { // push operation
				cache.conn.Write(append(append([]byte(ppCacheOp.Key + "-<-"), ppCacheOp.Data...), []byte("\r")...))
			}
		case cacheReply := <-cacheListener:
			for _, req := range ppCacheOpBuffer {
				if cacheReply.Key == req.Key {
					req.ReturnPayload <- cacheReply.Data
				}
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

	// waiting for request to be processed and retrieval of payload
	for !reply {
		select {
		case liveDataRes := <-request.ReturnPayload:
			fmt.Println("adasd")
			payload = liveDataRes
			reply = true
			break
		}
	}

	request = nil
	return payload
}
