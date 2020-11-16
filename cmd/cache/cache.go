package cache

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte
	Fullfilled bool
}

type Cache struct {
	cacheMap map[string][]byte
	bool CacheHandlerStarted
}

func New() Cache {
  cache := Cache{make(map[string][]byte)}
	cache.CacheHandlerStarted = false
	go StartCacheHandler()
  return cache
}

func (cache Cache) AddKeyVal(key string, val []byte) {
  cache.cacheMap[key] = val
}

func (cache Cache) StartCacheHandler(req <-chan PushPullRequest) {
	cache.CacheHandlerStarted = true
	requestBuffer := make([1000]*PushPullRequest)
	go func() {
		for {
			select {
			case request := <-req:
				append(cache.requestBuffer, &request)
			}
		}
	}()

	for request := range *requestBuffer {
			request.Fullfilled = true
			delete(cache.requestBuffer, &request)
	}
}
