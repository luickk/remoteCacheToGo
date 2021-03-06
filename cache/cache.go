package cache

import (
	"net"
	"bufio"
	"strconv"
	"errors"
	"crypto/tls"

  "remoteCacheToGo/pkg/util"
)

// struct to handle any requests for the CacheHandler which operates on cache memory
type PushPullRequest struct {
	Key string
	Operation string

	ReturnPayload chan []byte
	Data []byte

	ClientWriter *bufio.Writer
}

// describes a single client write request
// exists to forward client write requests form the handlers to the clientWrite handler via a channel
type clientWriteRequest struct {
	receivingClient *bufio.Writer
	data []byte
}

// stores all important data for cache
type Cache struct {
	// actual cache, holding all the data in memory
	cacheMem map[string][]byte

	PushPullRequestCh chan *PushPullRequest
	subscribedClients map[*bufio.Writer]int
	clientWriteRequestCh chan *clientWriteRequest
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048

// handles all requests to actual memory operations an the cache map
// aka. pushPullRequestHandler
func (cache Cache) CacheHandler(errorStream chan error) {
	var (
		err error
		encodedPPR []byte
	)
	encodingPPR := new(util.SPushPullReq)
	for {
		select {
		case ppCacheOp := <-cache.PushPullRequestCh:
			switch ppCacheOp.Operation {
			case ">":
				// checking if data attribute contains any data
				// if it doesn't no data needs to be pushed
				if len(ppCacheOp.Data) <= 0 { // pull operation
					// checking if cache-map contains key
					if _, ok := cache.cacheMem[ppCacheOp.Key]; ok {
						ppCacheOp.ReturnPayload <- cache.cacheMem[ppCacheOp.Key]
					} else {
						ppCacheOp.ReturnPayload <- []byte{}
					}
				// if it does, given data is written to key loc
				} else if len(ppCacheOp.Data) > 0 { // push operation
					for writer, _ := range cache.subscribedClients {
						encodingPPR.Key = ppCacheOp.Key
						encodingPPR.Operation = ">s"
						encodingPPR.Data = ppCacheOp.Data
						encodedPPR, err = util.EncodeMsg(encodingPPR)
						if err != nil {
							errorStream <- err
							return
						}
						cache.clientWriteRequestCh <- &clientWriteRequest { writer, encodedPPR }
					}

					cache.cacheMem[ppCacheOp.Key] = ppCacheOp.Data
				}
			case ">s":
				// reply to pull-request from chacheClient by index
				cache.subscribedClients[ppCacheOp.ClientWriter] = 0
			case ">s-c":
				// reply to pull-request from chacheClient by index
				delete(cache.subscribedClients, ppCacheOp.ClientWriter)
			}
		}
	}
}


// client handler, handles connected client sessions
// parses all incoming data
// hanles all operations on connection object
func (cache Cache)clientHandler(c net.Conn, errorStream chan error) {
	var (
		encodedPPR []byte
		err error
	)
	writer := bufio.NewWriter(c)
	decodedPPR := new(util.SPushPullReq)
	encodingPPR := new(util.SPushPullReq)
	netDataBuffer := make([]byte, tcpConnBuffer)
	for {
		netDataBuffer, err = util.ReadFrame(c)
		if err != nil {
			// write subscribed-client-list(subscribedClients) remove instruction
			cache.PushPullRequestCh <- &PushPullRequest{ "", ">s-c", nil, nil, writer }
			errorStream <- err
			return
		}
		// parsing instrucitons from client
		if err := util.DecodeMsg(decodedPPR, netDataBuffer); err != nil {
			return
		}
		switch decodedPPR.Operation {
		case ">":
			encodingPPR.Key = decodedPPR.Key
			encodingPPR.Operation = ">"
			encodingPPR.Data = cache.GetValByKey(decodedPPR.Key)
			if err != nil {
				errorStream <- err
				return
			}
			encodedPPR, err = util.EncodeMsg(encodingPPR)
			if err != nil {
				errorStream <- err
				return
			}
			cache.clientWriteRequestCh <- &clientWriteRequest { writer, encodedPPR }
		case ">s":
			cache.PushPullRequestCh <- &PushPullRequest{ "", ">s", nil, nil, writer }
		case "<":
			// writing push-request from client to cache
			cache.AddValByKey(decodedPPR.Key, decodedPPR.Data)
		}
	}
}

// provides network interface for given cache
func (cache Cache) RemoteConnHandler(bindAddress string, port int, errorStream chan error) {
	// opening tcp server
	l, err := net.Listen("tcp4", bindAddress+":"+strconv.Itoa(port))
	if err != nil {
		errorStream <- err
		return
	}
	defer l.Close()

	for {
		// waiting for client to connect
		c, err := l.Accept()
		if err != nil {
			errorStream <- err
			return
		}
		cache.clientHandler(c, errorStream)
	}
	return
}

// clientWriteRequestHandler handles all write request to clients
func (cache Cache)clientWriteRequestHandler(errorStream chan error) {
	for {
		writeRequest := <-cache.clientWriteRequestCh
		// err already handled at read
		util.WriteFrame(writeRequest.receivingClient, writeRequest.data)
	}
}

// provides network interface for given cache
// provides TLS encryption and password authentication
// provides valid and signed public/ private key pair and password hash to validate against
// parameters: port, password Hash (please don't use unhashed pw strings) enables delay between reconnects by ip, server Certificate, private Key
func (cache Cache) RemoteTlsConnHandler(port int, bindAddress string, token string, serverCert string, serverKey string, errorStream chan error) {
	// initiating provided key pair
	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		errorStream <- err
		return
	}

	// initiating config form key pair
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	// listening for clients who want to connect
	l, err := tls.Listen("tcp", bindAddress+":"+strconv.Itoa(port), config)
	if err != nil {
		errorStream <- err
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		tokenBuffer := make([]byte, 1024)
		if err != nil {
			errorStream <- err
			return
		}
    n, err := c.Read(tokenBuffer)
    if err != nil {
			errorStream <- err
			return
    }
		if string(tokenBuffer[:n]) == token {
			go cache.clientHandler(c, errorStream)
		} else {
			errorStream <- errors.New("token invalid")
			c.Close()
			return
		}
	}
}

// initiating new cache struct
func New(errorStream chan error) Cache {
	// initiating cache struct
  cache := Cache{ make(map[string][]byte), make(chan *PushPullRequest), make(map[*bufio.Writer]int), make(chan *clientWriteRequest)}

	// starting cache handler to allow for concurrent memory(cache map) operations
	go cache.CacheHandler(errorStream)

	go cache.clientWriteRequestHandler(errorStream)

  return cache
}


// adds key value to remote cache
// (can also overwrite/ replace)
func (cache Cache) AddValByKey(key string, val []byte) {
	// initiates push request
	// pushing request to pushPull handler
  cache.PushPullRequestCh <- &PushPullRequest{ key, ">", nil, val, nil }
}

// creates pull request for the remoteCache instance
func (cache Cache) GetValByKey(key string) []byte {
	// initiating pull request
	request := &PushPullRequest{ key, ">", make(chan []byte), nil, nil}
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
	return payload
}
