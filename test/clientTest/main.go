package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheClient"
)

func main() {
  fmt.Println("Client test")

  client := cacheClient.New("127.0.0.1", 8000)

  client.AddKeyVal("remote", []byte("test1"))
  fmt.Println("Written val test1 to key remote")

  fmt.Println("Read val from key remote: "+string(client.GetKeyVal("remote")))
}
