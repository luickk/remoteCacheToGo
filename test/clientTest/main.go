package main

import (
  "fmt"
  "strconv"
  "time"
  "remoteCacheToGo/cmd/cacheClient"
)

func main() {
  fmt.Println("Client test")

  client, err := cacheClient.New("127.0.0.1", 8000)
  if err != nil {
    fmt.Println(err)
    return
  }

  client.AddKeyVal("remote", []byte("test1"))
  fmt.Println("Written val test1 to key remote")

  fmt.Println("Read val from key remote: "+string(client.GetKeyVal("remote")))
  fmt.Println(": "+string(client.GetKeyVal("remotesada")))

  go concurrentTestInstanceA(client)
  concurrentTestInstanceB(client)
}
func concurrentTestInstanceA(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    client.AddKeyVal("remote"+strconv.Itoa(i), []byte("remote"+strconv.Itoa(i)))
    time.Sleep(1 * time.Millisecond)
  }
}

func concurrentTestInstanceB(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    fmt.Println("remote"+strconv.Itoa(i) + ": " + string(client.GetKeyVal("remote"+strconv.Itoa(i))))
    time.Sleep(10 * time.Millisecond)
    }
}
