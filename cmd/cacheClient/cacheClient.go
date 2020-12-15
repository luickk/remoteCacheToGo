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
	"os"
	"log"

  "remoteCacheToGo/pkg/util"
)

// defining different levels of loggers
var (
    WarningLogger *log.Logger
    InfoLogger    *log.Logger
    ErrorLogger   *log.Logger
)

// struct to handle any requests for the pushPullRequestHandler which operates on remoteCache connection
type PushPullRequest struct {
	Key string
	QueueIndex int
	ByIndex bool
	IsCountRequest bool

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
    WarningLogger.Println(err)
    return false, c
  }
  return true, c
}

// connectes to remoteCache and returns connection via. TLS protocol
// tls requires signed cert and password for authentication
func connectToTlsRemoteHandler(address string, port int, pwHash string, rootCert string) (bool, net.Conn) {
	// creating and appending new cert pool with x509 standard
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCert))
	if !ok {
		WarningLogger.Println("failed to parse root certificate")
	}
	// initing conifiguration for TLS connection
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


// connection-handler waits for incoming data
// incoming data is parsed and possible request answers are pushed to back to the pushPullRequestHandler
func connHandler(conn net.Conn, cacheRequestReply chan *PushPullRequest) {
	for {
		// initiating data buffer
		netData := make([]byte, tcpConnBuffer)
		// initiating reader
		n, err := bufio.NewReader(conn).Read(netData)
		if err != nil {
			WarningLogger.Println(err)
			return
		}
		// setting data slice to readers index
		netData = netData[:n]
		// splitting data to prevent overflow confusion
		netDataSeperated := bytes.Split(netData, []byte("\rnr"))
		if err != nil {
			WarningLogger.Println(err)
		}
		var dataDelimSplitByte [][]byte
		var key string
		var operation string
		var payload []byte
		// iterating over received data split by delimiter
		for _, data := range netDataSeperated {
			if len(data) >= 1 {
					// parsing protocol (you can find more about the protocol design in the README)
					dataDelimSplitByte = bytes.SplitN(data, []byte("-"), 3)
					// checking if protocol is valid, respectively completely available
					if len(dataDelimSplitByte) >= 3 {
						key = string(dataDelimSplitByte[0])
						operation = string(dataDelimSplitByte[1])
						payload = dataDelimSplitByte[2]
						// checks if incoming request wants to answer pull request
						if operation == ">" {
							// initiates reply to made pullrequest
							request := new(PushPullRequest)
							request.Key = key
							request.ByIndex = false
							request.IsCountRequest = false
							request.Data = payload
							cacheRequestReply <- request
							// tells gc it's ready to be removed
							request = nil
						} else if operation == ">i" {
							// initiates reply to made pullrequest
							request := new(PushPullRequest)
							request.ByIndex = true
							request.IsCountRequest = false
							request.QueueIndex, _ = strconv.Atoi(key)
							request.Data = payload
							cacheRequestReply <- request
							// tells gc it's ready to be removed
							request = nil
						} else if operation == ">c" {
							// initiates reply to made pullrequest
							request := new(PushPullRequest)
							request.ByIndex = false
							request.IsCountRequest = true
							request.QueueIndex, _ = strconv.Atoi(key)
							request.Data = payload
							cacheRequestReply <- request
							// tells gc it's ready to be removed
							request = nil
						}
					}
			}
		}
	}
}

// handles incoming requests on connected remoteCache
// reads/ writes form/ to connected cache
func (cache RemoteCache) pushPullRequestHandler() {
	// setting indicator for cache handler state to true
  cache.CacheHandlerStarted = true
	cacheRequestReply := make(chan *PushPullRequest)
	// push pull buffer, buffers all made pull requests to remoteCache instance to find & route replies to call
	var ppCacheOpBuffer []*PushPullRequest

	// starting connection Handler routine to parse incoming data and add to push request-replies to cacheListiner
	go connHandler(cache.conn, cacheRequestReply)

	for {
		select {
		// waits for push pull requests for remoteCache
		case ppCacheOp := <-cache.PushPullRequestCh:
			// checking if the request is a request for the current count of the cache
			if !ppCacheOp.IsCountRequest {
				// checking if the request is a request for a val by index
				if !ppCacheOp.ByIndex {
					if len(ppCacheOp.Data) <= 0 { // pull operation
						// adding request to cache-operation-buffer to assign it later to incoming request-reply
						ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
						// sends pull request string to remoteCache instance
						cache.conn.Write(append([]byte(ppCacheOp.Key + "->-"), []byte("\rnr")...))
					} else if len(ppCacheOp.Data) > 0 { // push operation
						// sends push request string to remoteCache instance
						cache.conn.Write(append(append([]byte(ppCacheOp.Key + "-<-"), ppCacheOp.Data...), []byte("\rnr")...))
					}
				// by index req
				} else {
					// adding request to cache-operation-buffer to assign it later to incoming request-reply
					ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
					// sending request to remoteCache instance
					cache.conn.Write(append([]byte(strconv.Itoa(ppCacheOp.QueueIndex) + "->i-"), []byte("\rnr")...))
				}
			// count req (by index)
			} else {
				// adding request to cache-operation-buffer to assign it later to incoming request-reply
				ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
				// sending request to remoteCache instance
				cache.conn.Write(append([]byte(strconv.Itoa(ppCacheOp.QueueIndex) + "->c-"), []byte("\rnr")...))
			}
		// waits for possible request replies to pull-requests from remoteCache instance via. channel from connection-handler
		case cacheReply := <-cacheRequestReply:
			for _, req := range ppCacheOpBuffer {
				// checks if both, buffered req and incoming req from remoteCache instance are not of "by-index type"
				if !cacheReply.ByIndex && !req.ByIndex {
					// compares buffered-requests with req. replies from remoteCache instance
					if cacheReply.Key == req.Key  && !req.Processed {
						// fullfills pull-requests data return
						req.ReturnPayload <- cacheReply.Data
						req.Processed = true
						
					} else if !req.Processed {
						// if request is not answered immeadiatly, request is forgotten
						req.Processed = true
					}
				// checks if both, buffered req and incoming req from remoteCache instance are of "by-index type"
				} else if cacheReply.ByIndex && req.ByIndex {
					// compares made requests from client with replies from remote cache
					if cacheReply.QueueIndex == req.QueueIndex  && !req.Processed {
						// fullfills pull-requests data return
						req.ReturnPayload <- cacheReply.Data
						req.Processed = true
					} else if !req.Processed {
						// if request is not answered immeadiatly, request is forgotten
						req.Processed = true
					}
				// checks if both, buffered req and incoming req from remoteCache instance are requests for the count
				} else if cacheReply.IsCountRequest && req.IsCountRequest {
					// compares made requests from client with replies from remote cache
					if cacheReply.QueueIndex == req.QueueIndex  && !req.Processed {
						// fullfills pull-requests data return
						req.ReturnPayload <- cacheReply.Data
						req.Processed = true
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
	// initing connection indicator and conn
	var connected bool
	var conn net.Conn

	InfoLogger = log.New(os.Stdout, "[NODE] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "[NODE] WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "[NODE] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// checking ig tls is enabled or not
	if tls {
		connected, conn = connectToTlsRemoteHandler(address, port, pwHash, rootCert)
	} else {
		connected, conn = connectToRemoteHandler(address, port)
	}
	// initing remote cache struct
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
func (cache RemoteCache) AddValByKey(key string, val []byte) bool {
  if util.CharacterWhiteList(key) {
		// initiates push request
		request := new(PushPullRequest)
		request.ByIndex = false
		request.IsCountRequest = false
		request.Key = key
		request.Data = val

	  cache.PushPullRequestCh <- request

		request = nil
		return true
	}
	return false
}

// creates pull request for the remoteCache instance
func (cache RemoteCache) GetValByKey(key string) []byte {
	if util.CharacterWhiteList(key) {
		// initiating pull request
		request := new(PushPullRequest)
		request.Key = key
		request.ByIndex = false
		request.IsCountRequest = false
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
func (cache RemoteCache) GetCountByIndex(index int) int {
	// initiating pull request
	request := new(PushPullRequest)
	request.QueueIndex = index
	request.ByIndex = false
	request.IsCountRequest = true

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
	// converting payload to string to int (payload = count)
	count, err := strconv.Atoi(string(payload))
	if err != nil {
		fmt.Println(err)
	}
	return count
}

// creates pull request for the remoteCache instance
func (cache RemoteCache) GetValByIndex(index int) []byte {
	// initiating pull request
	request := new(PushPullRequest)
	request.QueueIndex = index
	request.ByIndex = true
	request.IsCountRequest = false
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
