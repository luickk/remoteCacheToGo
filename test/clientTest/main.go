package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheClient"
)

func main() {
  fmt.Println("Client test")

  client := cacheClient.New("127.0.0.1", 8000)

  client.AddKeyVal("remote1", []byte("test1"))

  fmt.Println(string(client.GetKeyVal("remote1")))
}
