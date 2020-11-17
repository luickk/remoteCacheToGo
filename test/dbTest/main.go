package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheDb"
)

func main() {
	fmt.Println("DB test")

  cacheDb := cacheDb.New()
  cacheDb.NewCache("test")

  cacheDb.AddEntryToCache("test", "testKey", []byte("peter ist peter!"))


  fmt.Print(string(cacheDb.GetEntryFromCache("test", "testKey")))
}
