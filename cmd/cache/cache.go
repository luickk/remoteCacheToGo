package cache

import (
	"net"
	"log"
	"bufio"
	"bytes"
	"strconv"
	"time"
	"strings"
	"fmt"
	"crypto/tls"
	"os"

  "remoteCacheToGo/pkg/util"
  "remoteCacheToGo/pkg/goDosProtection"
)

// struct to handle any requests for the CacheHandler which operates on cache memory
type PushPullRequest struct {
	Key string
	QueueIndex int
	ByIndex bool
	ByIndexKey bool
	IsCountRequest bool

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
			if !ppCacheOp.IsCountRequest {
				// is a request for the value by key
				if !ppCacheOp.ByIndex && !ppCacheOp.ByIndexKey {
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
				// is a request for the value by index
				} else if ppCacheOp.ByIndex  {
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
				} else if ppCacheOp.ByIndexKey  {
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
				}
			// is a request for the count by index
			} else {
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

		// client handler, handles connected client sessions
		// parses all incoming data
		// hanles all operations on connection object
		go func(c net.Conn, cache Cache) {
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
				var dataDelimSplitByte [][]byte
				var key string
				var operation string
				var payload []byte
				// iterating over data seperated by delimiter
				for _, data := range netDataSeperated {
					if len(data) >= 1 {
							// parsing protocol (you can find more about the protocol design in the README)
							dataDelimSplitByte = bytes.SplitN(data, []byte("-"), 3)
							if len(dataDelimSplitByte) >= 3 {
								// protocol key (first element when seperated by "-" delim)
								key = string(dataDelimSplitByte[0])
								operation = string(dataDelimSplitByte[1])
								payload = dataDelimSplitByte[2]
								if operation == ">" { //pull by key
									// reply to pull-request from chacheClient by key
									c.Write(append(append([]byte(key+"->-"), cache.GetValByKey(key)...), []byte("\rnr")...))
								} else if operation == ">i" { //pull by index
									index, err := strconv.Atoi(key)
									if err != nil {
										WarningLogger.Println(err)
									}
									// reply to pull-request from chacheClient by index
									c.Write(append(append([]byte(key+"->i-"), cache.GetValByIndex(index)...), []byte("\rnr")...))
								} else if operation == ">ik" { //pull by index
									index, err := strconv.Atoi(key)
									if err != nil {
										WarningLogger.Println(err)
									}
									fmt.Println(cache.GetKeyByIndex(index))
									// reply to pull-request from chacheClient by index
									c.Write(append(append([]byte(key+"->ik-"), []byte(cache.GetKeyByIndex(index))...), []byte("\rnr")...))
								} else if operation == ">c" { //pull by index
									index, err := strconv.Atoi(key)
									if err != nil {
										WarningLogger.Println(err)
									}
									// reply to pull-request from chacheClient by index
									c.Write(append(append([]byte(key+"->c-"), []byte(strconv.Itoa(cache.GetCountByIndex(index)))...), []byte("\rnr")...))
								} else if operation == "<" { // push
									// writing push-request from client to cache
									cache.AddValByKey(key, payload)
								}
							} else {
									WarningLogger.Println("Parsing Error")
							}
						}
					}
				}
		}(c, cache)
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

			// connection listener waits for incoming data
			// incoming data is parsed and possible request answers are pushed to back to the CacheHandler
			go func(c net.Conn, cache Cache) {
				InfoLogger.Println("New client connected to %s \n", c.RemoteAddr().String())
				InfoLogger.Println("waiting for authentication")
				// true if client sent correct password hash
				var authenticated = false
				// limits authentification tries
				var bruteForceTimer = false
				// time each client has to wait before next reconnect
				var bruteForceProtectionTime = 1

				var dataDelimSplitByte [][]byte
				var key string
				var operation string
				var payload []byte
				for {
					// data buffer
					netData := make([]byte, tcpConnBuffer)
					n, err := bufio.NewReader(c).Read(netData)
					if err != nil {
						ErrorLogger.Println(err)
						return
					}
					// extracting data read by reader from buffer
					netData = netData[:n]

					// splitting data to prevent overflow confusion
					netDataSeperated := bytes.Split(netData, []byte("\rnr"))
					if err != nil {
						WarningLogger.Println(err)
						return
					}

					// splitting data to prevent buffer overflow confusion
					for _, data := range netDataSeperated {
						if len(data) >= 1 {
							// checking if client has authenticated
							if authenticated {
								// parsing protocol (you can find more about the protocol design in the README)
								dataDelimSplitByte = bytes.SplitN(data, []byte("-"), 3)
								if len(dataDelimSplitByte) >= 3 {
									// protocol key (first element when seperated by "-" delim)
									key = string(dataDelimSplitByte[0])
									operation = string(dataDelimSplitByte[1])
									payload = dataDelimSplitByte[2]
									// if request operation is pull, the pull request is replied
									if operation == ">" { //pull
										// replying to pull request with requested data
										c.Write(append(append([]byte(key+"->-"), cache.GetValByKey(key)...), []byte("\rnr")...))
										// executing push request
									} else if operation == "<" { // push
										// setting value for given key
										cache.AddValByKey(key, payload)
									}
								} else {
									WarningLogger.Println("Parsing error")
								}
							} else {
								// checking if password is valid and if the bruteForceTimer is finished
								if string(data) == pwHash  && !bruteForceTimer {
									InfoLogger.Println("Authentification successful")
									authenticated = true
								// bruteForceTimer has not finished yet
								} else if bruteForceTimer {
									InfoLogger.Println("Client tried to authenticate in brute force protection time")
								// password was invalid
								} else {
									InfoLogger.Println("Authentification unsuccessful")
									bruteForceTimer = true
									// resetting timer
									timer := time.NewTimer(time.Second*time.Duration(bruteForceProtectionTime))
									go func() {
										<-timer.C
										bruteForceTimer = false
							    }()
								}
							}
						}
					}
				}
			}(c, cache)
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
		request.ByIndex = false
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
	request.ByIndex = true
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
	request.ByIndex = false
	request.ByIndexKey = true
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
