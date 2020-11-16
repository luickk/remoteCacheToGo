package cache

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte

	Fullfilled bool
}

type Cache struct {
	cacheMap map[string][]byte
	CacheHandlerStarted bool
}

func (cache Cache) StartCacheHandler(req<-chan PushPullRequest) {
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

	for request := range requestBuffer {
			request.Fullfilled = true
			// pull operation
			if len(request.Data) <= 0 {
				request.ReturnPayload <- cache.cacheMap[request.Key]
			} else if len(request.Data) > 0 { // push operation
				cache.cacheMap[request.Key] = request.Data
			}
			delete(requestBuffer, &request)
	}
}

func New() Cache {
  cache := Cache{make(map[string][]byte), false}
	cache.CacheHandlerStarted = false
	go StartCacheHandler()
  return cache
}

func (cache Cache) AddKeyVal(key string, val []byte) {
  cache.cacheMap[key] = val
}
