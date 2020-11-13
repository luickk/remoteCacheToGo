package serverInstance

import (
	"bytes"
  "remoteCacheToGo/cmd/cache"
)

type cacheDb struct {
	dbs map[string][cache]
}

func New() {
  cacheDb := cacheDb{make(map[string][cache])}
  return cacheDb
}
