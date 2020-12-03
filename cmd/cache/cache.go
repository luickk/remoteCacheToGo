package cache

import (
	"net"
	"fmt"
	"bufio"
	"bytes"
	"strconv"
	"crypto/tls"
  "remoteCacheToGo/internal/remoteCacheToGo"
)

// struct to handle any requests for the CacheHandler which operates on cache memory
type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte
}

// stores all important data for cache
type Cache struct {
	// actual cache holding all data in single cache
	cacheMap map[string][]byte
	PushPullRequestCh chan *PushPullRequest
	CacheHandlerStarted bool
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048

// handles all requests to actual memory operations an the cache map
func (cache Cache) CacheHandler() {
	cache.CacheHandlerStarted = true
	for {
		select {
		case ppCacheOp := <-cache.PushPullRequestCh:
			// checking if data attribute contains
			// if it doesn't no data needs to be pushed
			if len(ppCacheOp.Data) <= 0 { // pull operation
				ppCacheOp.ReturnPayload <- cache.cacheMap[ppCacheOp.Key]
			// if it does, given data is written to key
			} else if len(ppCacheOp.Data) > 0 { // push operation
				cache.cacheMap[ppCacheOp.Key] = ppCacheOp.Data
			}
		}
	}
}

// provides network interface for given cache
func (cache Cache) RemoteConnHandler(port int) {
	// opening tcp server
	l, err := net.Listen("tcp4", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		// waiting for client to connect
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		// client handler, handles connected client sessions
		// parses all incoming data
		// hanles all operations on connection object
		go func(c net.Conn, cache Cache) {
			fmt.Printf("New client connected to %s \n", c.RemoteAddr().String())
			for {
				data := make([]byte, tcpConnBuffer)
				n, err := bufio.NewReader(c).Read(data)
				if err != nil {
					fmt.Println(err)
					return
				}
				data = data[:n]

				// splitting data to prevent buffer overflow confusion
				netDataSeperated := bytes.Split(data, []byte("\rnr"))
				if err != nil {
					fmt.Println(err)
					return
				}

				for _, data := range netDataSeperated {
					// fmt.Println(strconv.Itoa(len(data)) + ": " + string(data))
					if len(data) >= 1 {
							// parsing protocol (you can find more about the protocol design in the README)
							dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 3)
							if len(dataDelimSplitByte) >= 3 {
								key := string(dataDelimSplitByte[0])
								operation := string(dataDelimSplitByte[1])
								payload := dataDelimSplitByte[2]
								if operation == ">" { //pull
									// reply to pull request from chacheClient
									c.Write(append(append([]byte(key+"->-"), cache.GetKeyVal(key)...), []byte("\rnr")...))
								} else if operation == "<" { // push
									// writing push request from client to cache
									cache.AddKeyVal(key, payload)
								}
							} else {
									fmt.Println("parsing error")
							}
						}
					}
				}
		}(c, cache)
	}
}

// provides network interface for given cache
// provide TLS encryption and password authentication
// requres valid and signed public/ private key pair and password hash to validate against
func (cache Cache) RemoteTlsConnHandler(port int, pwHash string, serverCert string, serverKey string) {
	// initiating provided key pair
	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		fmt.Println(err)
	}
	// initiating config form key pair
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	// listening for clients who want to connect
	l, err := tls.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port), config)
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		// connection listener waits for incoming data
		// incoming data is parsed and possible request answers are pushed to back to the CacheHandler
		go func(c net.Conn, cache Cache) {
			fmt.Println("New client connected to %s \n", c.RemoteAddr().String())
			fmt.Println("waiting for authentication")
			// true if client sent correct password hash
			var authenticated = false
			for {
				data := make([]byte, tcpConnBuffer)
				n, err := bufio.NewReader(c).Read(data)
				if err != nil {
					fmt.Println(err)
					return
				}
				data = data[:n]

				// splitting data to prevent overflow confusion
				netDataSeperated := bytes.Split(data, []byte("\rnr"))
				if err != nil {
					fmt.Println(err)
					return
				}

				for _, data := range netDataSeperated {
					if len(data) >= 1 {
						if authenticated {
							// parsing protocol (you can find more about the protocol design in the README)
							dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 3)
							if len(dataDelimSplitByte) >= 3 {
								key := string(dataDelimSplitByte[0])
								operation := string(dataDelimSplitByte[1])
								payload := dataDelimSplitByte[2]
								// if request operation is pull, the pull request is replied
								if operation == ">" { //pull
									// replying to pull request with requested data
									c.Write(append(append([]byte(key+"->-"), cache.GetKeyVal(key)...), []byte("\rnr")...))
									// executing push request
								} else if operation == "<" { // push
									// setting value for given key
									cache.AddKeyVal(key, payload)
								}
							} else {
								fmt.Println("parsing error")
							}
						} else {
							if string(data) == pwHash {
								fmt.Println("Authentification successful")
								authenticated = true
							} else {
								fmt.Println("Authentification unsuccessful")
							}
						}
					}
				}
			}
		}(c, cache)
	}
}

// initiating new cache struct
func New() Cache {
  cache := Cache{make(map[string][]byte), make(chan *PushPullRequest), false}
	cache.CacheHandlerStarted = false
	// starting cache handler to allow for concurrent memory(cache map) operations
	go cache.CacheHandler()
  return cache
}


// adds key value to remote cache
// (can also overwrite/ replace)
func (cache Cache) AddKeyVal(key string, val []byte) bool {
	// checking if inserting key data contains unallowed characters
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
func (cache Cache) GetKeyVal(key string) []byte {
	// checking if inserting key data contains unallowed characters to block unnecessary request
  if util.CharacterWhiteList(key) {
		// initiating pull request
		request := new(PushPullRequest)
		request.Key = key
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
	return []byte{}
}
