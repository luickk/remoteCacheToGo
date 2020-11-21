package cache

import (
	"net"
)

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte
}

type Cache struct {
	cacheMap map[string][]byte
	PushPullRequestCh chan *PushPullRequest
	CacheHandlerStarted bool
}

func (cache Cache) CacheHandler(req chan *PushPullRequest) {
	cache.CacheHandlerStarted = true
	for {
		select {
		case ppCacheOp := <-req:
			// pull operation
			if len(ppCacheOp.Data) <= 0 {
				ppCacheOp.ReturnPayload <- cache.cacheMap[ppCacheOp.Key]
			} else if len(ppCacheOp.Data) > 0 { // push operation
				cache.cacheMap[ppCacheOp.Key] = ppCacheOp.Data
			}
		}
	}
}

func (cache Cache) RemoteConnHandler(port int) {
	l, err := net.Listen("tcp4", port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go func handleConnection(c net.Conn) {
			fmt.Printf("New client connected to %s \n", c.RemoteAddr().String())
			for {
				netData, err := bufio.NewReader(c).ReadString('\r')
				if err != nil {
					fmt.Println(err)
					return
				}
				if netData[0] == ">" { //pull

				} else if netData[0] == "<" { // push

				}
			}
			result := strconv.Itoa(random()) + "\n"
			c.Write([]byte(string(result)))
		}(c)
	}
}

func New() Cache {
  cache := Cache{make(map[string][]byte), make(chan *PushPullRequest), false}
	cache.CacheHandlerStarted = false
	go cache.CacheHandler(cache.PushPullRequestCh)
  return cache
}

func (cache Cache) AddKeyVal(key string, val []byte) {
	request := new(PushPullRequest)
	request.Key = key
	request.Data = val

  cache.PushPullRequestCh <- request

	request = nil
}

func (cache Cache) GetKeyVal(key string) []byte {
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
