package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheDb"
  "strconv"
  "time"
)

func main() {
	fmt.Println("DB test")

  cDb := cacheDb.New()
  cDb.NewCache("test")

  if cDb.AddEntryToCache("test", "testKey", []byte("test1")) {
    fmt.Println("Written val to testKey")
  }

  fmt.Println("Requestd key: "+string(cDb.GetEntryFromCache("test", "testKey")))

  cDb.Db["test"].RemoteConnHandler(8000)

  //
  // go concurrentTestInstanceA(cDb)
  // concurrentTestInstanceB(cDb)
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
    time.Sleep(10 * time.Millisecond)
    }
}
