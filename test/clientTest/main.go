package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheClient"
)

func main() {
  fmt.Println("Client test")

  cacheClient.New("localhost", 444)

  // cDb.AddEntryToCache("test1", "testKey", []byte("test1"))

  // fmt.Print(string(cDb.GetEntryFromCache("test1", "testKey")))
}
