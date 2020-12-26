package cache

import (
	"net"
	"log"
	"bufio"
	"bytes"
	"strconv"
	"strings"
	"crypto/tls"
	"os"

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
}

// stores all important data for cache
type Cache struct {
	// actual cache, holding all the data in memory
	cacheMem map[string]*CacheVal
	PushPullRequestCh chan *PushPullRequest
	CacheHandlerStarted bool
}

// required since the queueIndex is also in the cache map key
type CacheVal struct {
	Data []byte
	QueueIndex int
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048

// defining different levels of loggers
var (
    WarningLogger *log.Logger
    InfoLogger    *log.Logger
    ErrorLogger   *log.Logger
)

// handles all requests to actual memory operations an the cache map
// aka. pushPullRequestHandler
func (cache Cache) CacheHandler() {
	var queueIndex int = 1
	cache.CacheHandlerStarted = true
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
					val := new(CacheVal)
					val.QueueIndex = queueIndex
					val.Data = ppCacheOp.Data
					cache.cacheMem[ppCacheOp.Key] = val
					// increasing queueIndex (count)
					queueIndex++
				}
			case ">i":
				var found bool
				var reversedReqIndex int
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
				var found bool
				var reversedReqIndex int
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
				var found bool
				var reversedReqIndex int
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
			}
		}
	}
}


// client handler, handles connected client sessions
// parses all incoming data
// hanles all operations on connection object
func clientHandler(c net.Conn, cache Cache) {
	InfoLogger.Printf("New client connected to %s \n", c.RemoteAddr().String())
	for {
		netData := make([]byte, tcpConnBuffer)
		n, err := bufio.NewReader(c).Read(netData)
		if err != nil {
			ErrorLogger.Println(err)
			return
		}
		netData = netData[:n]

		// splitting data to prevent buffer overflow confusion
		netDataSeperated := bytes.Split(netData, []byte("\rnr"))
		if err != nil {
			WarningLogger.Println(err)
			return
		}

		var encodedPPR []byte
		decodedPPR := new(util.SPushPullReq)
		encodingPPR := new(util.SPushPullReq)
		// iterating over data seperated by delimiter
		for _, data := range netDataSeperated {
			if len(data) >= 1 {
						if err := util.DecodePushPullReq(decodedPPR, data); err != nil {
							WarningLogger.Println(err)
						}
						switch decodedPPR.Operation {
						case ">":
							encodingPPR.Key = decodedPPR.Key
							encodingPPR.Operation = ">"
							encodingPPR.Data = cache.GetValByKey(decodedPPR.Key)
							encodedPPR, err = util.EncodePushPullReq(encodingPPR)
							if err != nil {
								WarningLogger.Println(err)
							}
							// reply to pull-request from chacheClient by key
							c.Write(append(encodedPPR, []byte("\rnr")...))
						case ">i":
							index, err := strconv.Atoi(decodedPPR.Key)
							if err != nil {
								WarningLogger.Println(err)
							}
							// reply to pull-request from chacheClient by index
							encodingPPR.Key = decodedPPR.Key
							encodingPPR.Operation = ">i"
							encodingPPR.Data = cache.GetValByIndex(index)
							encodedPPR, err = util.EncodePushPullReq(encodingPPR)
							if err != nil {
								WarningLogger.Println(err)
							}
							// reply to pull-request from chacheClient by key
							c.Write(append(encodedPPR, []byte("\rnr")...))
						case ">ik":
							index, err := strconv.Atoi(decodedPPR.Key)
							if err != nil {
								WarningLogger.Println(err)
							}
							// reply to pull-request from chacheClient by index
							encodingPPR.Key = decodedPPR.Key
							encodingPPR.Operation = ">ik"
							encodingPPR.Data = []byte(cache.GetKeyByIndex(index))
							encodedPPR, err = util.EncodePushPullReq(encodingPPR)
							if err != nil {
								WarningLogger.Println(err)
							}
							// reply to pull-request from chacheClient by key
							c.Write(append(encodedPPR, []byte("\rnr")...))
						case ">c":
							index, err := strconv.Atoi(decodedPPR.Key)
							if err != nil {
								WarningLogger.Println(err)
							}
							// reply to pull-request from chacheClient by index
							encodingPPR.Key = decodedPPR.Key
							encodingPPR.Operation = ">c"
							encodingPPR.Data = []byte(strconv.Itoa(cache.GetCountByIndex(index)))
							encodedPPR, err = util.EncodePushPullReq(encodingPPR)
							if err != nil {
								WarningLogger.Println(err)
							}
							// reply to pull-request from chacheClient by key
							c.Write(append(encodedPPR, []byte("\rnr")...))
						case "<":
							// writing push-request from client to cache
							cache.AddValByKey(decodedPPR.Key, decodedPPR.Data)
						default:
								WarningLogger.Println("Parsing Error")
						}
				}
			}
		}
}

// provides network interface for given cache
func (cache Cache) RemoteConnHandler(port int) {
	// opening tcp server
	l, err := net.Listen("tcp4", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		ErrorLogger.Println(err)
		return
	}
	defer l.Close()

	for {
		// waiting for client to connect
		c, err := l.Accept()
		if err != nil {
			ErrorLogger.Println(err)
			return
		}
		go clientHandler(c, cache)
	}
}

// provides network interface for given cache
// provides TLS encryption and password authentication
// provides valid and signed public/ private key pair and password hash to validate against
// parameters: port, password Hash (please don't use unhashed pw strings), dosProtection enables delay between reconnects by ip, server Certificate, private Key
func (cache Cache) RemoteTlsConnHandler(port int, pwHash string, dosProtection bool, serverCert string, serverKey string) {
	// initiating provided key pair
	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		ErrorLogger.Println(err)
	}

	// initiating DOS protection with 10 second reconnection delay
  dosProt := goDosProtection.New(10)

	// initiating config form key pair
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	// listening for clients who want to connect
	l, err := tls.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port), config)
	if err != nil {
		ErrorLogger.Println(err)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			ErrorLogger.Println(err)
			return
		}

		// client is not banned
		if !dosProt.Client(strings.Split(c.RemoteAddr().String(), ":")[0]) || !dosProtection {
		  InfoLogger.Println("Accepted client connection")
			go clientHandler(c, cache)
		// client is banned
		} else {
		 InfoLogger.Println("Refused client connection")
		}
	}
}

// initiating new cache struct
func New() Cache {
	// initiating cache struct
  cache := Cache{ make(map[string]*CacheVal), make(chan *PushPullRequest), false }
	// updating cacheHandler state
	cache.CacheHandlerStarted = false

	// initiating different logger
	InfoLogger = log.New(os.Stdout, "[NODE] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "[NODE] WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "[NODE] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// starting cache handler to allow for concurrent memory(cache map) operations
	go cache.CacheHandler()
  return cache
}


// adds key value to remote cache
// (can also overwrite/ replace)
func (cache Cache) AddValByKey(key string, val []byte) bool {
	// checking if inserting key data contains unallowed characters
  if util.CharacterWhiteList(key) {
		// initiates push request
		request := new(PushPullRequest)
		request.Key = key
		request.Data = val
		request.Operation = ">"
		// pushing request to pushPull handler
	  cache.PushPullRequestCh <- request

		request = nil
		return true
	}
	return false
}

// creates pull request for the remoteCache instance
func (cache Cache) GetValByKey(key string) []byte {
	// checking if inserting key data contains unallowed characters to block unnecessary request
  if util.CharacterWhiteList(key) {
		// initiating pull request
		request := new(PushPullRequest)
		request.Key = key
		request.ReturnPayload = make(chan []byte)
		request.Operation = ">"
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
func (cache Cache) GetCountByIndex(index int) int {
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
	var count int
	var err error
	if len(payload) != 0 {
		count, err = strconv.Atoi(string(payload))
		if err != nil {
			// queueindex begins at 1
			WarningLogger.Println(err)
		}
	} else {
		return 0
	}

	return count
}

// creates pull request for the remoteCache instance
func (cache Cache) GetValByIndex(index int) []byte {
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
func (cache Cache) GetKeyByIndex(index int) string {
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
