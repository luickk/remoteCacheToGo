package cache

import (
	"net"
	"fmt"
	"bufio"
	"bytes"
	"strconv"
  "remoteCacheToGo/internal/remoteCacheToGo"
)

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte
}

type Cache struct {
	cacheMap map[string][]byte
	PushPullRequestCh chan *PushPullRequest
	CacheHandlerStarted bool
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 1024

func (cache Cache) CacheHandler() {
	cache.CacheHandlerStarted = true
	for {
		select {
		case ppCacheOp := <-cache.PushPullRequestCh:
			if len(ppCacheOp.Data) <= 0 { // pull operation
				ppCacheOp.ReturnPayload <- cache.cacheMap[ppCacheOp.Key]
			} else if len(ppCacheOp.Data) > 0 { // push operation
				cache.cacheMap[ppCacheOp.Key] = ppCacheOp.Data
			}
		}
	}
}

func (cache Cache) RemoteConnHandler(port int) {
	l, err := net.Listen("tcp4", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go func(c net.Conn, cache Cache) {
			fmt.Printf("New client connected to %s \n", c.RemoteAddr().String())
			data := make([]byte, tcpConnBuffer)
			for {
				n, err := bufio.NewReader(c).Read(data)
				if err != nil {
					fmt.Println(err)
					return
				}
				data = data[:n]

				netDataSeperated := bytes.Split(data, []byte("\r"))
				if err != nil {
					fmt.Println(err)
					return
				}

				for _, data := range netDataSeperated {
					if len(data) >= 1 {
							dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 3)
							if len(dataDelimSplitByte) >= 3 {	
								key := string(dataDelimSplitByte[0])
								operation := string(dataDelimSplitByte[1])
								payload := dataDelimSplitByte[2]
								if operation == ">" { //pull
									c.Write(append(append([]byte(key+"->-"), cache.GetKeyVal(key)...), []byte("\r")...))
								} else if operation == "<" { // push
									cache.AddKeyVal(key, payload)
								}
							}
						}
					}
				}
		}(c, cache)
	}
}

func New() Cache {
  cache := Cache{make(map[string][]byte), make(chan *PushPullRequest), false}
	cache.CacheHandlerStarted = false
	go cache.CacheHandler()
  return cache
}

func (cache Cache) AddKeyVal(key string, val []byte) bool {
  if util.CharacterWhiteList(key) {
		request := new(PushPullRequest)
		request.Key = key
		request.Data = val

	  cache.PushPullRequestCh <- request

		request = nil
		return true
	}
	return false
}

func (cache Cache) GetKeyVal(key string) []byte {
  if util.CharacterWhiteList(key) {
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
	return []byte{}
}
