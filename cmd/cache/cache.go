package cache

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte

	Fullfilled bool
}

type Cache struct {
	cacheMap map[string][]byte
	PushPullRequestCh chan PushPullRequest
	CacheHandlerStarted bool
}

func (cache Cache) CacheHandler(req chan PushPullRequest) {
	cache.CacheHandlerStarted = true
	var requestBuffer []PushPullRequest

	go func(rB *[]PushPullRequest) {
		for {
			select {
			case request := <-req:
				*rB = append(*rB, request)
			}
		}
	}(&requestBuffer)

	for {
		for _, request := range requestBuffer {
			request.Fullfilled = true
			// pull operation
			if len(request.Data) <= 0 {
				request.ReturnPayload <- cache.cacheMap[request.Key]
			} else if len(request.Data) > 0 { // push operation
				cache.cacheMap[request.Key] = request.Data
			}
		}
	}
}

func New() Cache {
  cache := Cache{make(map[string][]byte), make(chan PushPullRequest), false}
	cache.CacheHandlerStarted = false
	go cache.CacheHandler(cache.PushPullRequestCh)
  return cache
}

func (cache Cache) AddKeyVal(key string, val []byte) {
	request := new(PushPullRequest)
	request.Key = key
	request.Data = val
	request.Fullfilled = false

  cache.PushPullRequestCh <- *request
}

func (cache Cache) GetKeyVal(key string) []byte {
	request := new(PushPullRequest)
	request.Key = key
	request.Fullfilled = false
	request.ReturnPayload = make(chan []byte)
  cache.PushPullRequestCh <- *request

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
	return payload
}
