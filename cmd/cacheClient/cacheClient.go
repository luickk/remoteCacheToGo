package cacheClient

import (
	"net"
	"fmt"
	"strconv"
	"bytes"
	"bufio"
	"crypto/tls"
	"crypto/x509"
  "errors"
  "remoteCacheToGo/pkg/util"
)

// struct to handle any requests for the pushPullRequestHandler which operates on remoteCache connection
type PushPullRequest struct {
	Key string
	QueueIndex int
	ByIndex bool

	ReturnPayload chan []byte
	Data []byte
	Processed bool
}

// acts as interface to talk to remoteCache instance (if conected successfully)
type RemoteCache struct {
	conn net.Conn
	PushPullRequestCh chan *PushPullRequest
	CacheHandlerStarted bool
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048

// connectes to remoteCache and returns connection
// NO tls NO authentication
func connectToRemoteHandler(address string, port int) (bool, net.Conn) {
	// dialing unencrypted tcp connection
  c, err := net.Dial("tcp", address+":"+strconv.Itoa(port))
  if err != nil {
    fmt.Println(err)
    return false, c
  }
  return true, c
}

// connectes to remoteCache and returns connection via. TLS protocol
// tls requires signed cert and password for authentication
func connectToTlsRemoteHandler(address string, port int, pwHash string, rootCert string) (bool, net.Conn) {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCert))
	if !ok {
		fmt.Println("failed to parse root certificate")
	}
	config := &tls.Config{RootCAs: roots}

	// dial encrypted tls connection
	c, err := tls.Dial("tcp", address+":"+strconv.Itoa(port), config)
	if err != nil {
		fmt.Println(err)
    return false, c
	}
	// sending password hash in order to authenticate
	c.Write([]byte(pwHash))

  return true, c
}

// handles incoming requests on connected remoteCache
// reads/ writes form/ to connected cache
func (cache RemoteCache) pushPullRequestHandler() {
  cache.CacheHandlerStarted = true
	cacheListener := make(chan *PushPullRequest)
	// push pull buffer, buffers all made pull requests to remoteCache instance to find & route replies to call
	var ppCacheOpBuffer []*PushPullRequest

	// connection listener waits for incoming data
	// incoming data is parsed and possible request answers are pushed to back to the pushPullRequestHandler
	go func(conn net.Conn, cacheListener chan *PushPullRequest) {
		for {
			// initiating data buffer
			data := make([]byte, tcpConnBuffer)
			// initiating reader
			n, err := bufio.NewReader(conn).Read(data)
			if err != nil {
				fmt.Println(err)
				return
			}
			data = data[:n]
			// splitting data to prevent overflow confusion
			netDataSeperated := bytes.Split(data, []byte("\rnr"))
			if err != nil {
				fmt.Println(err)
			}

			for _, data := range netDataSeperated {
				if len(data) >= 1 {
						// parsing protocol (you can find more about the protocol design in the README)
						dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 3)
						if len(dataDelimSplitByte) >= 3 {
							key := string(dataDelimSplitByte[0])
							operation := string(dataDelimSplitByte[1])
							payload := dataDelimSplitByte[2]
							// checks if incoming request wants to answer pull request
							if operation == ">" {
								// initiates reply to made pullrequest
								request := new(PushPullRequest)
								request.Key = key
								request.ByIndex = false
								request.Data = payload
								cacheListener <- request
								request = nil
							} else if operation == ">i" {
								fmt.Println("receved")
								// initiates reply to made pullrequest
								request := new(PushPullRequest)
								request.ByIndex = true
								request.QueueIndex, _ = strconv.Atoi(key)
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
		// waits for push pull requests for remoteCache
		case ppCacheOp := <-cache.PushPullRequestCh:
			if !ppCacheOp.ByIndex {
				if len(ppCacheOp.Data) <= 0 { // pull operation
					ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
					// sends pull request string to remoteCache instance
					cache.conn.Write(append([]byte(ppCacheOp.Key + "->-"), []byte("\rnr")...))
				} else if len(ppCacheOp.Data) > 0 { // push operation
				// sends push request string to remoteCache instance
					cache.conn.Write(append(append([]byte(ppCacheOp.Key + "-<-"), ppCacheOp.Data...), []byte("\rnr")...))
				}
			} else {
				cache.conn.Write(append([]byte(strconv.Itoa(ppCacheOp.QueueIndex) + "->i-"), []byte("\rnr")...))
			}
		// waits for possible reqplies to pull requests from remoteCache
		case cacheReply := <-cacheListener:
			for _, req := range ppCacheOpBuffer {
				if !cacheReply.ByIndex && !req.ByIndex {
					// compares made requests from client with replies from remote cache
					if cacheReply.Key == req.Key  && !req.Processed {
						// fullfills pull requests data return
						req.ReturnPayload <- cacheReply.Data
					} else if !req.Processed {
						// if request is not answered immeadiatly, request is forgotten
						req.Processed = true
					}
				} else if cacheReply.ByIndex && req.ByIndex {
					// compares made requests from client with replies from remote cache
					if cacheReply.QueueIndex == req.QueueIndex  && !req.Processed {
						// fullfills pull requests data return
						req.ReturnPayload <- cacheReply.Data
					} else if !req.Processed {
						// if request is not answered immeadiatly, request is forgotten
						req.Processed = true
					}
				}
			}
		}
	}
}

// initiates new RemoteCache struct and connects to remoteCache instance
// params concerning tls (tls, pwHash, rootCert) can be initiated empty if tls bool is false
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
	// starts pushPullRequestHandler for concurrent request handling
  go cache.pushPullRequestHandler()
	return cache, nil
}

// adds key value to remote cache
// (can also overwrite/ replace)
func (cache RemoteCache) AddKeyVal(key string, val []byte) bool {
  if util.CharacterWhiteList(key) {
		// initiates push request
		request := new(PushPullRequest)
		request.Key = key
		request.Data = val

	  cache.PushPullRequestCh <- request

		request = nil
		return true
	}
	return false
}

// creates pull request for the remoteCache instance
func (cache RemoteCache) GetKeyVal(key string) []byte {
	if util.CharacterWhiteList(key) {
		// initiating pull request
		request := new(PushPullRequest)
		request.Key = key
		request.ByIndex = false
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


// creates pull request for the remoteCache instance
func (cache RemoteCache) GetIndexVal(index int) []byte {
	// initiating pull request
	request := new(PushPullRequest)
	request.QueueIndex = index
	request.ByIndex = true
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
