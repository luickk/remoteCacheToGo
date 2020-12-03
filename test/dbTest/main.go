package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheDb"
  "strconv"
  "time"
)

func main() {
	fmt.Println("DB test")

  // creating new cache database
  cDb := cacheDb.New()

  // creating new caches in database
  cDb.NewCache("test")
  cDb.NewCache("remote")

  // adding new entry to cache "test" at key "testkey" with val "test1"
  if cDb.AddEntryToCache("test", "testKey", []byte("test1")) {
    fmt.Println("Written val to testKey")
  }

  // pulling data from cache "test" at key "testKey"
  fmt.Println("Requestd key: "+string(cDb.GetEntryFromCache("test", "testKey")))

  // creating unencrypted network interfce for cache with name "remote"
  cDb.Db["remote"].RemoteConnHandler(8000)

  // runing test instances
  go concurrentTestInstanceA(cDb)
  concurrentTestInstanceB(cDb)
}

func concurrentTestInstanceA(cDb cacheDb.CacheDb) {
  i := 0
  for {
    i++
    cDb.AddEntryToCache("test", "test"+strconv.Itoa(i), []byte("test"+strconv.Itoa(i)))
  }
}

func concurrentTestInstanceB(cDb cacheDb.CacheDb) {
  i := 0
  for {
    i++
    fmt.Println("test"+strconv.Itoa(i) + ": " + string(cDb.GetEntryFromCache("test", "test"+strconv.Itoa(i))))
    fmt.Println("remote: "+string(cDb.GetEntryFromCache("remote", "remote")))
    time.Sleep(10 * time.Millisecond)
    }
}
