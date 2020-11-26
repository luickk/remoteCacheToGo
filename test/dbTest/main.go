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

  cDb.Db["test"].RemoteConnHandler(8000)

  // if cDb.AddEntryToCache("test1", "testKey", []byte("test1")) {
  //   fmt.Println("Written val to testKey")
  // }

  // fmt.Println("Requestd key: "+string(cDb.GetEntryFromCache("test1", "testKey")))

  // go concurrentTestInstanceA(cDb)
  // concurrentTestInstanceB(cDb)
}

func concurrentTestInstanceA(cDb cacheDb.CacheDb) {
  i := 0
  for {
    i++
    cDb.AddEntryToCache("test2", "test"+strconv.Itoa(i), []byte("test"+strconv.Itoa(i)))
    cDb.AddEntryToCache("test1", "test"+strconv.Itoa(i), []byte("test"+strconv.Itoa(i)))
  }
}

func concurrentTestInstanceB(cDb cacheDb.CacheDb) {
  i := 0
  for {
    i++
    fmt.Println("test"+strconv.Itoa(i) + ": " + string(cDb.GetEntryFromCache("test2", "test"+strconv.Itoa(i))))
    fmt.Println("test"+strconv.Itoa(i) + ": " + string(cDb.GetEntryFromCache("test1", "test"+strconv.Itoa(i))))
    time.Sleep(10 * time.Millisecond)
    }
}
