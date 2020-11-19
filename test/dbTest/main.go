package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheDb"
)

func main() {
	fmt.Println("DB test")

  cDb := cacheDb.New()
  cDb.NewCache("test1")
  cDb.NewCache("test2")

  cDb.AddEntryToCache("test1", "testKey", []byte("test1"))

  fmt.Print(string(cDb.GetEntryFromCache("test1", "testKey")))

  go concurrentTestInstanceA(cDb)
  concurrentTestInstanceB(cDb)
}

func concurrentTestInstanceA(cDb cacheDb.CacheDb) {
  i := 0
  for {
    cDb.AddEntryToCache("test2", "test"+string(i), []byte("test2"))
    fmt.Print(cDb.GetEntryFromCache("test1", "test"+string(i)))
  }
}

func concurrentTestInstanceB(cDb cacheDb.CacheDb) {
  i := 0
  for {
    cDb.AddEntryToCache("test1", "test"+string(i), []byte("test1"))
    fmt.Print(cDb.GetEntryFromCache("test2", "test"+string(i)))

  }
}
