package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheClient"
)

func main() {
  fmt.Println("Client test")

  client := cacheClient.New("localhost", 8000)

  client.AddKeyVal("remote1", []byte("test1"))

  fmt.Println(string(client.GetKeyVal("remote1")))
}
