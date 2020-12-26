package cacheClient

import (
	"net"
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
	Operation string

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
		ErrorLogger.Println("Failed to parse root certificate")
	}
	// initing conifiguration for TLS connection
	config := &tls.Config{RootCAs: roots}

	// dial encrypted tls connection
	c, err := tls.Dial("tcp", address+":"+strconv.Itoa(port), config)
	if err != nil {
		ErrorLogger.Println(err)
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

		decodedPPR := new(util.SPushPullReq)

		// iterating over received data split by delimiter
		for _, data := range netDataSeperated {
			if len(data) >= 1 {
						if err := util.DecodePushPullReq(decodedPPR, data); err != nil {
							WarningLogger.Println(err)
						}
						switch decodedPPR.Operation {
						case ">":
							// initiates reply to made pullrequest
							request := new(PushPullRequest)
							request.Key = decodedPPR.Key
							request.Operation = ">"
							request.Data = decodedPPR.Data
							cacheRequestReply <- request
							// tells gc it's ready to be removed
							request = nil
						case ">i":
							// initiates reply to made pullrequest
							request := new(PushPullRequest)
							request.Operation = ">i"
							request.QueueIndex, _ = strconv.Atoi(decodedPPR.Key)
							request.Data = decodedPPR.Data
							cacheRequestReply <- request
							// tells gc it's ready to be removed
							request = nil
						case ">ik":
							// initiates reply to made pullrequest
							request := new(PushPullRequest)
							request.Operation = ">ik"
							request.QueueIndex, _ = strconv.Atoi(decodedPPR.Key)
							request.Data = decodedPPR.Data
							cacheRequestReply <- request
							// tells gc it's ready to be removed
							request = nil
						case ">c":
							// initiates reply to made pullrequest
							request := new(PushPullRequest)
							request.Operation = ">c"
							request.QueueIndex, _ = strconv.Atoi(decodedPPR.Key)
							request.Data = decodedPPR.Data
							cacheRequestReply <- request
							// tells gc it's ready to be removed
							request = nil
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
	var encodedPPR []byte
	var err error
	encodingPPR := new(util.SPushPullReq)

	// starting connection Handler routine to parse incoming data and add to push request-replies to cacheListiner
	go connHandler(cache.conn, cacheRequestReply)

	for {
		select {
		// waits for push pull requests for remoteCache
		case ppCacheOp := <-cache.PushPullRequestCh:
			switch ppCacheOp.Operation {
			case ">":
				if len(ppCacheOp.Data) <= 0 { // pull operation
					// adding request to cache-operation-buffer to assign it later to incoming request-reply
					ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
					// sends pull request string to remoteCache instance
					encodingPPR.Key = ppCacheOp.Key
					encodingPPR.Operation = ">"
					encodingPPR.Data = []byte{}
					encodedPPR, err = util.EncodePushPullReq(encodingPPR)
					if err != nil {
						WarningLogger.Println(err)
					}
					// reply to pull-request from chacheClient by key
					cache.conn.Write(append(encodedPPR, []byte("\rnr")...))
					} else if len(ppCacheOp.Data) > 0 { // push operation
					// sends push request string to remoteCache instance
					encodingPPR.Key = ppCacheOp.Key
					encodingPPR.Operation = "<"
					encodingPPR.Data = ppCacheOp.Data
					encodedPPR, err = util.EncodePushPullReq(encodingPPR)
					if err != nil {
						WarningLogger.Println(err)
					}
					// reply to pull-request from chacheClient by key
					cache.conn.Write(append(encodedPPR, []byte("\rnr")...))
				}
			case ">i":
				// adding request to cache-operation-buffer to assign it later to incoming request-reply
				ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
				// sending request to remoteCache instance
				encodingPPR.Key = strconv.Itoa(ppCacheOp.QueueIndex)
				encodingPPR.Operation = ">i"
				encodingPPR.Data = []byte{}
				encodedPPR, err = util.EncodePushPullReq(encodingPPR)
				if err != nil {
					WarningLogger.Println(err)
				}
				// reply to pull-request from chacheClient by key
				cache.conn.Write(append(encodedPPR, []byte("\rnr")...))
			case ">ik":
				// adding request to cache-operation-buffer to assign it later to incoming request-reply
				ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
				// sending request to remoteCache instance
				encodingPPR.Key = strconv.Itoa(ppCacheOp.QueueIndex)
				encodingPPR.Operation = ">ik"
				encodingPPR.Data = []byte{}
				encodedPPR, err = util.EncodePushPullReq(encodingPPR)
				if err != nil {
					WarningLogger.Println(err)
				}
				// reply to pull-request from chacheClient by key
				cache.conn.Write(append(encodedPPR, []byte("\rnr")...))
			case ">c":
				// adding request to cache-operation-buffer to assign it later to incoming request-reply
				ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
				// sending request to remoteCache instance
				encodingPPR.Key = strconv.Itoa(ppCacheOp.QueueIndex)
				encodingPPR.Operation = ">c"
				encodingPPR.Data = []byte{}
				encodedPPR, err = util.EncodePushPullReq(encodingPPR)
				if err != nil {
					WarningLogger.Println(err)
				}
				// reply to pull-request from chacheClient by key
				cache.conn.Write(append(encodedPPR, []byte("\rnr")...))
			}
		// waits for possible request replies to pull-requests from remoteCache instance via. channel from connection-handler
		case cacheReply := <-cacheRequestReply:
			for _, req := range ppCacheOpBuffer {
				switch (cacheReply.Operation + req.Operation) {
				case ">>":
					if !req.Processed {
						// fullfills pull-requests data return
						req.ReturnPayload <- cacheReply.Data
						req.Processed = true
					} else if !req.Processed {
						// if request is not answered immeadiatly, request is forgotten
						req.Processed = true
					}
				case ">i>i":
					if !req.Processed {
						// fullfills pull-requests data return
						req.ReturnPayload <- cacheReply.Data
						req.Processed = true

					} else if !req.Processed {
						// if request is not answered immeadiatly, request is forgotten
						req.Processed = true
					}
				case ">ik>ik":
					// compares made requests from client with replies from remote cache
					if !req.Processed {
						// fullfills pull-requests data return
						req.ReturnPayload <- cacheReply.Data
						req.Processed = true
					} else if !req.Processed {
						// if request is not answered immeadiatly, request is forgotten
						req.Processed = true
					}
				case ">c>c":
					// compares made requests from client with replies from remote cache
					if !req.Processed {
						// fullfills pull-requests data return
						req.ReturnPayload <- cacheReply.Data
						req.Processed = true
					} else if !req.Processed {
						// if request is not answered immeadiatly, request is forgotten
						req.Processed = true
					}
				}
				// removing all ppOp from ppCacheOpBuffer if processed
				ppCacheOpBuffer = removeOperation(ppCacheOpBuffer, util.Index(len(ppCacheOpBuffer), func(i int) bool { return ppCacheOpBuffer[i].Processed }))
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

	// initiating different logger
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
		request.Operation = ">"
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
		request.Operation = ">"
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
	request.Operation = ">c"

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
		WarningLogger.Println(err)
	}
	return count
}

// creates pull request for the remoteCache instance
func (cache RemoteCache) GetValByIndex(index int) []byte {
	// initiating pull request
	request := new(PushPullRequest)
	request.QueueIndex = index
	request.Operation = ">i"
	request.ReturnPayload = make(chan []byte)
  cache.PushPullRequestCh <- request

	var reply bool
	var payload []byte

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

// creates pull request for the remoteCache instance
func (cache RemoteCache) GetKeyByIndex(index int) string {
	// initiating pull request
	request := new(PushPullRequest)
	request.QueueIndex = index
	request.Operation = ">ik"
	request.ReturnPayload = make(chan []byte)
  cache.PushPullRequestCh <- request

	var reply bool
	var payload string

	// waiting for request to be processed and retrieval of payload
	for !reply {
		select {
		case liveDataRes := <-request.ReturnPayload:
			payload = string(liveDataRes)
			reply = true
			break
		}
	}
	request = nil
	return payload
}

// ppOp slice operation
// only use if order is not of importance
func removeOperation(s []*PushPullRequest, i int) []*PushPullRequest {
    s[i] = s[len(s)-1]
    // We do not need to put s[i] at the end, as it will be discarded anyway
    return s[:len(s)-1]
}
