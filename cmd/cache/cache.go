package cache

import (
	"net"
	"log"
	"bufio"
	"errors"
	"strconv"
	"strings"
	"crypto/tls"
	"os"
	"time"

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
}

// required since the queueIndex is also in the cache map key
type CacheVal struct {
	Data []byte
	QueueIndex int
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048
var maxKeySize = 500
var maxValSize = 500

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
	time.Sleep(1*time.Millisecond)
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

	var encodedPPR []byte
	var err error
	decodedPPR := new(util.SPushPullReq)
	encodingPPR := new(util.SPushPullReq)
	reader := bufio.NewReader(c)
	netDataBuffer := make([]byte, tcpConnBuffer)
	for {
		netDataBuffer, err = reader.ReadBytes(0x0)
		if err != nil {
			WarningLogger.Println(err)
			break
		}
		netDataBuffer = netDataBuffer[:len(netDataBuffer)-1]
		if len(netDataBuffer) > 0 {
			if err = util.DecodePushPullReq(decodedPPR, netDataBuffer); err != nil {
				WarningLogger.Println(err)
			}
			switch decodedPPR.Operation {
			case ">":
				encodingPPR.Key = decodedPPR.Key
				encodingPPR.Operation = ">"
				encodingPPR.Data, err = cache.GetValByKey(decodedPPR.Key)
				if err != nil {
					WarningLogger.Println(err)
				}
				encodedPPR, err = util.EncodePushPullReq(encodingPPR)
				if err != nil {
					WarningLogger.Println(err)
				}
				c.Write(append(encodedPPR, []byte{0x0}...))
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
				c.Write(append(encodedPPR, []byte{0x0}...))
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
				c.Write(append(encodedPPR, []byte{0x0}...))
			case ">c":
				index, err := strconv.Atoi(decodedPPR.Key)
				if err != nil {
					WarningLogger.Println(err)
				}
				// reply to pull-request from chacheClient by index
				encodingPPR.Key = decodedPPR.Key
				encodingPPR.Operation = ">c"
				count, err := cache.GetCountByIndex(index)
				if err != nil {
					WarningLogger.Println(err)
				}
				encodingPPR.Data = []byte(strconv.Itoa(count))
				encodedPPR, err = util.EncodePushPullReq(encodingPPR)
				if err != nil {
					WarningLogger.Println(err)
				}
				c.Write(append(encodedPPR, []byte{0x0}...))
			case "<":
				// writing push-request from client to cache
				if err := cache.AddValByKey(decodedPPR.Key, decodedPPR.Data); err != nil {
					WarningLogger.Println(err)
				}
			default:
					WarningLogger.Println("Parsing Error")
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
  cache := Cache{ make(map[string]*CacheVal), make(chan *PushPullRequest) }

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
func (cache Cache) AddValByKey(key string, val []byte) error {
  if len(key) < maxKeySize && len(val) < maxValSize {
		// initiates push request
		request := new(PushPullRequest)
		request.Key = key
		request.Data = val
		request.Operation = ">"
		// pushing request to pushPull handler
	  cache.PushPullRequestCh <- request

		request = nil
		return nil
	}
	return errors.New("key or val size exceeds limit")
}

// creates pull request for the remoteCache instance
func (cache Cache) GetValByKey(key string) ([]byte, error) {
  if len(key) < maxKeySize {
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
		return payload, nil
	}
	return []byte{}, errors.New("key size exceeds limit")
}

// creates pull request for the remoteCache instance
func (cache Cache) GetCountByIndex(index int) (int, error) {
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
			return 0, err
		}
	} else {
		return 0, nil
	}
	return count, nil
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
