package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheDb"
)

func main() {
	fmt.Println("DB test")

  cacheDb := cacheDb.New()
  cacheDb.NewCache("test")
  fmt.Println("1")
  cacheDb.AddEntryToCache("test", "testKey", []byte("peter ist peter!"))
  fmt.Println("2")

  fmt.Print(string(cacheDb.GetEntryFromCache("test", "testKey")))
  fmt.Println("3")
}
