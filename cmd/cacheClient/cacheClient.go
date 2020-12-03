package cacheClient

import (
	"net"
	"fmt"
	"strconv"
	"bytes"
	"bufio"
	"crypto/tls"
	"crypto/x509"
  "remoteCacheToGo/internal/remoteCacheToGo"
  "errors"
)

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte
	Processed bool
}

type RemoteCache struct {
	conn net.Conn
	PushPullRequestCh chan *PushPullRequest
	CacheHandlerStarted bool
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048

func connectToRemoteHandler(address string, port int) (bool, net.Conn) {
  c, err := net.Dial("tcp", address+":"+strconv.Itoa(port))
  if err != nil {
    fmt.Println(err)
    return false, c
  }
  return true, c
}

func connectToTlsRemoteHandler(address string, port int, pwHash string, rootCert string) (bool, net.Conn) {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCert))
	if !ok {
		fmt.Println("failed to parse root certificate")
	}
	config := &tls.Config{RootCAs: roots}

	c, err := tls.Dial("tcp", address+":"+strconv.Itoa(port), config)
	if err != nil {
		fmt.Println(err)
    return false, c
	}
	c.Write([]byte(pwHash))
  return true, c
}


func (cache RemoteCache) pushPullRequestHandler() {
  cache.CacheHandlerStarted = true
	cacheListener := make(chan *PushPullRequest)
	var ppCacheOpBuffer []*PushPullRequest

	go func(conn net.Conn, cacheListener chan *PushPullRequest) {
		for {
			data := make([]byte, tcpConnBuffer)
			n, err := bufio.NewReader(conn).Read(data)
			if err != nil {
				fmt.Println(err)
				return
			}
			data = data[:n]
			netDataSeperated := bytes.Split(data, []byte("\rnr"))
			if err != nil {
				fmt.Println(err)
			}

			for _, data := range netDataSeperated {
				if len(data) >= 1 {
						dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 3)
						if len(dataDelimSplitByte) >= 3 {
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
		}
	}(cache.conn, cacheListener)

	for {
		select {
		case ppCacheOp := <-cache.PushPullRequestCh:
			if len(ppCacheOp.Data) <= 0 { // pull operation
				ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
				cache.conn.Write(append([]byte(ppCacheOp.Key + "->-"), []byte("\rnr")...))
			} else if len(ppCacheOp.Data) > 0 { // push operation
				cache.conn.Write(append(append([]byte(ppCacheOp.Key + "-<-"), ppCacheOp.Data...), []byte("\rnr")...))
			}
		case cacheReply := <-cacheListener:
			for _, req := range ppCacheOpBuffer {
				if cacheReply.Key == req.Key  && !req.Processed {
					req.ReturnPayload <- cacheReply.Data

				} else if !req.Processed {
					req.Processed = true
				}
			}
		}
	}
}

func New(address string, port int, tls bool, pwHash string, rootCert string) (RemoteCache, error) {
	var connected bool
	var conn net.Conn
	if tls {
		connected, conn = connectToTlsRemoteHandler(address, port, pwHash, rootCert)
	} else {
		connected, conn = connectToRemoteHandler(address, port)
	}
	cache := RemoteCache{conn, make(chan *PushPullRequest), false}
  if !connected {
		return cache, errors.New("timeout")
  }
  go cache.pushPullRequestHandler()
	return cache, nil
}

func (cache RemoteCache) AddKeyVal(key string, val []byte) bool {
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

func (cache RemoteCache) GetKeyVal(key string) []byte {
	if util.CharacterWhiteList(key) {
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
