PushPullRequest{package cacheClient

import (
	"net"
	"strconv"
	"bufio"
	"crypto/tls"
	"crypto/x509"
  "errors"

  "remoteCacheToGo/pkg/util"
)

type subscribeCacheVal struct {
	Key string
	Data []byte
}

// struct to handle any requests for the pushPullRequestHandler which operates on remoteCache connection
type PushPullRequest struct {
	Key string
	QueueIndex int
	Operation string

	ReturnPayload chan []byte
	Data []byte
	Processed bool

	SubscriptionReturn chan *subscribeCacheVal
}

// acts as interface to talk to remoteCache instance (if conected successfully)
type RemoteCache struct {
	conn net.Conn
	PushPullRequestCh chan PushPullRequest
	CacheHandlerStarted bool
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048

// connectes to remoteCache and returns connection
// NO tls NO authentication
func connectToRemoteHandler(address string, port int) (net.Conn, error) {
	// dialing unencrypted tcp connection
  c, err := net.Dial("tcp", address+":"+strconv.Itoa(port))
  if err != nil {
    return c, err
  }
  return c, nil
}

// connectes to remoteCache and returns connection via. TLS protocol
// tls requires signed cert and password for authentication
func connectToTlsRemoteHandler(address string, port int, pwHash string, rootCert string) (net.Conn, error) {
	var err error
	var c net.Conn
	// creating and appending new cert pool with x509 standard
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCert))
	if !ok {
		return c, errors.New("failed to parse root certificate")
	}
	// initing conifiguration for TLS connection
	config := &tls.Config{RootCAs: roots}

	// dial encrypted tls connection
	c, err = tls.Dial("tcp", address+":"+strconv.Itoa(port), config)
	if err != nil {
    return c, err
	}
	// sending password hash in order to authenticate
	c.Write([]byte(pwHash))

  return c, err
}


// connection-handler waits for incoming data
// incoming data is parsed and possible request answers are pushed to back to the pushPullRequestHandler
func connHandler(conn net.Conn, cacheRequestReply chan PushPullRequest) {
	var err error
	decodedPPR := new(util.SPushPullReq)
	netDataBuffer := make([]byte, tcpConnBuffer)
	request := new(PushPullRequest)

	for {
		netDataBuffer, err = util.ReadFrame(conn)
		if err != nil {
			return
		}
		// parsing instrucitons from client
		if err := util.DecodeMsg(decodedPPR, netDataBuffer); err != nil {
			return
		}
		switch decodedPPR.Operation {
		case ">":
			// initiates reply to made pullrequest
			request.Key = decodedPPR.Key
			request.Operation = ">"
			request.Data = decodedPPR.Data
			cacheRequestReply <- *request
		case ">i":
			request.Operation = ">i"
			request.QueueIndex, _ = strconv.Atoi(decodedPPR.Key)
			request.Data = decodedPPR.Data
			cacheRequestReply <- *request
		case ">ik":
			request.Operation = ">ik"
			request.QueueIndex, _ = strconv.Atoi(decodedPPR.Key)
			request.Data = decodedPPR.Data
			cacheRequestReply <- *request
		case ">c":
			request.Operation = ">c"
			request.QueueIndex, _ = strconv.Atoi(decodedPPR.Key)
			request.Data = decodedPPR.Data
			cacheRequestReply <- *request
		case ">s":
			request.Operation = ">s"
			request.Key = decodedPPR.Key
			request.Data = decodedPPR.Data
			cacheRequestReply <- *request
		}
	}
}

// handles incoming requests on connected remoteCache
// reads/ writes form/ to connected cache
func (cache RemoteCache) pushPullRequestHandler() {
	// setting indicator for cache handler state to true
  cache.CacheHandlerStarted = true
	cacheRequestReply := make(chan PushPullRequest)
	// push pull buffer, buffers all made pull requests to remoteCache instance to find & route replies to call
	var ppCacheOpBuffer []PushPullRequest
	var encodedPPR []byte
	var err error
	encodingPPR := new(util.SPushPullReq)
	writer := bufio.NewWriter(cache.conn)

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
						encodedPPR, err = util.EncodeMsg(encodingPPR)
						if err != nil {
							return
						}
						// reply to pull-request from chacheClient by key
						util.WriteFrame(writer, encodedPPR)
					} else if len(ppCacheOp.Data) > 0 { // push operation
						// sends push request string to remoteCache instance
						encodingPPR.Key = ppCacheOp.Key
						encodingPPR.Operation = "<"
						encodingPPR.Data = ppCacheOp.Data
						encodedPPR, err = util.EncodeMsg(encodingPPR)
						if err != nil {
							return
						}
						util.WriteFrame(writer, encodedPPR)
					}
				case ">i":
					// adding request to cache-operation-buffer to assign it later to incoming request-reply
					ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
					// sending request to remoteCache instance
					encodingPPR.Key = strconv.Itoa(ppCacheOp.QueueIndex)
					encodingPPR.Operation = ">i"
					encodingPPR.Data = []byte{}
					encodedPPR, err = util.EncodeMsg(encodingPPR)
					if err != nil {
						return
					}
					util.WriteFrame(writer, encodedPPR)
				case ">ik":
					// adding request to cache-operation-buffer to assign it later to incoming request-reply
					ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
					// sending request to remoteCache instance
					encodingPPR.Key = strconv.Itoa(ppCacheOp.QueueIndex)
					encodingPPR.Operation = ">ik"
					encodingPPR.Data = []byte{}
					encodedPPR, err = util.EncodeMsg(encodingPPR)
					if err != nil {
						return
					}
					util.WriteFrame(writer, encodedPPR)
				case ">c":
					// adding request to cache-operation-buffer to assign it later to incoming request-reply
					ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
					// sending request to remoteCache instance
					encodingPPR.Key = strconv.Itoa(ppCacheOp.QueueIndex)
					encodingPPR.Operation = ">c"
					encodingPPR.Data = []byte{}
					encodedPPR, err = util.EncodeMsg(encodingPPR)
					if err != nil {
						return
					}
					util.WriteFrame(writer, encodedPPR)
				case ">s":
					// adding request to cache-operation-buffer to assign it later to incoming request-reply
					ppCacheOpBuffer = append(ppCacheOpBuffer, ppCacheOp)
					// sending request to remoteCache instance
					encodingPPR.Operation = ">s"
					encodedPPR, err = util.EncodeMsg(encodingPPR)
					if err != nil {
						return
					}
					util.WriteFrame(writer, encodedPPR)
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
				case ">s>s":
					req.SubscriptionReturn <- &subscribeCacheVal{ cacheReply.Key, cacheReply.Data }
					req.Processed = false
				}
				if req.Processed {
					// removing all ppOp from ppCacheOpBuffer if processed
					ppCacheOpBuffer = removeOperation(ppCacheOpBuffer, util.Index(len(ppCacheOpBuffer), func(i int) bool { return ppCacheOpBuffer[i].Processed }))
				}
			}
		}
	}
}

// initiates new RemoteCache struct and connects to remoteCache instance
// params concerning tls (tls, pwHash, rootCert) can be initiated empty if tls bool is false
func New(address string, port int, tls bool, pwHash string, rootCert string) (RemoteCache, error) {
	var conn net.Conn
	var err error

	// initing remote cache struct
	cache := RemoteCache{conn, make(chan *PushPullRequest), false}

	// checking ig tls is enabled or not
	if tls {
		conn, err = connectToTlsRemoteHandler(address, port, pwHash, rootCert)
	  if err != nil {
			return cache, err
	  }
	} else {
		conn, err = connectToRemoteHandler(address, port)
	  if err != nil {
			return cache, err
	  }
	}
	cache.conn = conn

	// starts pushPullRequestHandler for concurrent request handling
  go cache.pushPullRequestHandler()
	return cache, nil
}

// adds key value to remote cache
// (can also overwrite/ replace)
func (cache RemoteCache) AddValByKey(key string, val []byte) error {
	if len(key) > 0 && len(val) > 0 {
		// initiates push request
	  cache.PushPullRequestCh <- PushPullRequest{ key, 0, ">", nil, val, false, nil }
		return nil
	}
	return errors.New("key or value are empty")
}

// creates pull request for the remoteCache instance
func (cache RemoteCache) GetValByKey(key string) ([]byte, error) {
	if len(key) > 0 {
		request := PushPullRequest{ key, 0, ">", make(chan []byte), nil, false, nil }
		// initiating pull request
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
		return payload, nil
	}
	return []byte{}, errors.New("key is empty")
}

// creates pull request for the remoteCache instance
func (cache RemoteCache) GetCountByIndex(index int) (int, error) {
	// initiating pull request
	request := PushPullRequest{ "", index, ">c", make(chan []byte), nil, false, nil }

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
		return 0, err
	}
	return count, nil
}

// creates pull request for the remoteCache instance
func (cache RemoteCache) GetValByIndex(index int) []byte {
	// initiating pull request
	request := PushPullRequest{ "", index, ">i", make(chan []byte), nil, false, nil }
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
	request := PushPullRequest{ "", index, ">ik", make(chan []byte), nil, false, nil }
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

// creates pull request for the remoteCache instance
func (cache RemoteCache) Subscribe() chan *subscribeCacheVal {
	// initiating pull request
	request := PushPullRequest{ "", 0, ">s", nil, nil, false, make(chan *subscribeCacheVal) }
	cache.PushPullRequestCh <- request

	return request.SubscriptionReturn
}

// ppOp slice operation
// only use if order is not of importance
func removeOperation(s []*PushPullRequest, i int) []*PushPullRequest {
    s[i] = s[len(s)-1]
    // We do not need to put s[i] at the end, as it will be discarded anyway
    return s[:len(s)-1]
}
