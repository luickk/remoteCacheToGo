package cache

import (
	"net"
	"bufio"
	"strconv"
	"strings"
	"crypto/tls"

  "remoteCacheToGo/pkg/util"
  "remoteCacheToGo/pkg/goDosProtection"
)

// struct to handle any requests for the CacheHandler which operates on cache memory
type PushPullRequest struct {
	Key string
	QueueIndex int
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
	cacheMem map[string]*CacheVal

	PushPullRequestCh chan *PushPullRequest
	subscribedClients map[*bufio.Writer]int
	clientWriteRequestCh chan *clientWriteRequest
}

// required since the queueIndex is also in the cache map key
type CacheVal struct {
	Data []byte
	QueueIndex int
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048

// handles all requests to actual memory operations an the cache map
// aka. pushPullRequestHandler
func (cache Cache) CacheHandler() error {
	var (
		err error
		queueIndex int = 1
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
						ppCacheOp.ReturnPayload <- cache.cacheMem[ppCacheOp.Key].Data
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
							return err
						}
						cache.clientWriteRequestCh <- &clientWriteRequest { writer, encodedPPR }
					}

					cache.cacheMem[ppCacheOp.Key] = &CacheVal{ ppCacheOp.Data, queueIndex }

					// increasing queueIndex (count)
					queueIndex++
				}
			case ">i":
				var (
					found bool
					reversedReqIndex int
				)
				// Iterate over cacheMem CacheVal which stores queue index
				for _, data := range cache.cacheMem {
					if !(ppCacheOp.QueueIndex > len(cache.cacheMem)) {
						// calculating reversed index
						reversedReqIndex = len(cache.cacheMem) - ppCacheOp.QueueIndex
					} else {
						// setting reversed index to max len of cache-map
						reversedReqIndex = len(cache.cacheMem)
					}
					if data.QueueIndex == reversedReqIndex {
						found = true
						// returning value at key to return channel
						ppCacheOp.ReturnPayload <- data.Data
					}
				}
				if !found {
					// returning zero bytes if if there is no matching index
					ppCacheOp.ReturnPayload <- []byte{}
				}
			case ">ik":
				var (
					found bool
					reversedReqIndex int
				)
				// Iterate over cacheMem CacheVal which stores queue index
				for key, data := range cache.cacheMem {
					if !(ppCacheOp.QueueIndex > len(cache.cacheMem)) {
						// calculating reversed index
						reversedReqIndex = len(cache.cacheMem) - ppCacheOp.QueueIndex
					} else {
						// setting reversed index to max len of cache-map
						reversedReqIndex = len(cache.cacheMem)
					}
					if data.QueueIndex == reversedReqIndex {
						found = true
						// returning value at key to return channel
						ppCacheOp.ReturnPayload <- []byte(key)
					}
				}
				if !found {
					// returning zero bytes if if there is no matching index
					ppCacheOp.ReturnPayload <- []byte{}
				}
			case ">c":
				var (
					found bool
					reversedReqIndex int
				)
				for _, data := range cache.cacheMem {
					// checking if request index is within the "index-space" of the cache-map
					if !(ppCacheOp.QueueIndex > len(cache.cacheMem)) {
						// calculating reversed index
						reversedReqIndex = len(cache.cacheMem) - ppCacheOp.QueueIndex
					} else {
						// setting reversed index to min len of cache-map
						reversedReqIndex = 1
					}
					// finding element with matching index
					if data.QueueIndex == reversedReqIndex {
						// returning value at index to return channel
						ppCacheOp.ReturnPayload <- []byte(strconv.Itoa(data.QueueIndex))
						found = true
					}
				}
				if !found {
					// returning zero bytes if if there is no matching index
					ppCacheOp.ReturnPayload <- []byte{}
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
func (cache Cache)clientHandler(c net.Conn) error {
	var encodedPPR []byte
	var err error
	writer := bufio.NewWriter(c)
	decodedPPR := new(util.SPushPullReq)
	encodingPPR := new(util.SPushPullReq)
	netDataBuffer := make([]byte, tcpConnBuffer)
	for {
			netDataBuffer, err = util.ReadFrame(c)
			if err != nil {
				// write subscribed-client-list(subscribedClients) remove instruction
				cache.PushPullRequestCh <- &PushPullRequest{ "", 0, ">s-c", nil, nil, writer }
				return err
			}
			// parsing instrucitons from client
			if err := util.DecodeMsg(decodedPPR, netDataBuffer); err != nil {
				return err
			}
		switch decodedPPR.Operation {
		case ">":
			encodingPPR.Key = decodedPPR.Key
			encodingPPR.Operation = ">"
			encodingPPR.Data = cache.GetValByKey(decodedPPR.Key)
			if err != nil {
				return err
			}
			encodedPPR, err = util.EncodeMsg(encodingPPR)
			if err != nil {
				return err
			}
			cache.clientWriteRequestCh <- &clientWriteRequest { writer, encodedPPR }
		case ">s":
			cache.PushPullRequestCh <- &PushPullRequest{ "", 0, ">s", nil, nil, writer }
		case "<":
			// writing push-request from client to cache
			cache.AddValByKey(decodedPPR.Key, decodedPPR.Data)
		default:
				return nil
		}
	}
}

// provides network interface for given cache
func (cache Cache) RemoteConnHandler(bindAddress string, port int) error {
	// opening tcp server
	l, err := net.Listen("tcp4", bindAddress+":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		// waiting for client to connect
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) error {
			return cache.clientHandler(conn)
		}(c)
	}
	return nil
}

// clientWriteRequestHandler handles all write request to clients
func (cache Cache)clientWriteRequestHandler() error {
	for {
		select {
		case writeRequest := <-cache.clientWriteRequestCh:
			if err := util.WriteFrame(writeRequest.receivingClient, writeRequest.data); err != nil {
				return err
			}
		}
	}
}

// provides network interface for given cache
// provides TLS encryption and password authentication
// provides valid and signed public/ private key pair and password hash to validate against
// parameters: port, password Hash (please don't use unhashed pw strings), dosProtection enables delay between reconnects by ip, server Certificate, private Key
func (cache Cache) RemoteTlsConnHandler(port int, bindAddress string, pwHash string, dosProtection bool, serverCert string, serverKey string) error {
	// initiating provided key pair
	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		return err
	}

	// initiating DOS protection with 10 second reconnection delay
  dosProt := goDosProtection.New(10)

	// initiating config form key pair
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	// listening for clients who want to connect
	l, err := tls.Listen("tcp", bindAddress+strconv.Itoa(port), config)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}

		// client is not banned
		if !dosProt.Client(strings.Split(c.RemoteAddr().String(), ":")[0]) || !dosProtection {
			go func(conn net.Conn) error {
				return cache.clientHandler(conn)
			}(c)
		// client is banned
		}
	}
}

// initiating new cache struct
func New() Cache {
	// initiating cache struct
  cache := Cache{ make(map[string]*CacheVal), make(chan *PushPullRequest), make(map[*bufio.Writer]int), make(chan *clientWriteRequest)}

	// starting cache handler to allow for concurrent memory(cache map) operations
	go func() error {
		return cache.CacheHandler()
	}()
	go func() error {
		return cache.clientWriteRequestHandler()
	}()
  return cache
}


// adds key value to remote cache
// (can also overwrite/ replace)
func (cache Cache) AddValByKey(key string, val []byte) {
	// initiates push request
	// pushing request to pushPull handler
  cache.PushPullRequestCh <- &PushPullRequest{ key, 0, ">", nil, val, nil }
}

// creates pull request for the remoteCache instance
func (cache Cache) GetValByKey(key string) []byte {
	// initiating pull request
	request := &PushPullRequest{ key, 0, ">", make(chan []byte), nil, nil}
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
