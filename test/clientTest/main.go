package main

import (
  "fmt"
  "strconv"
  "time"
  "remoteCacheToGo/cmd/cacheClient"
)

func main() {
  fmt.Println("Client test")

  // creates new cacheClient struct and connects to remoteCache instance
  // no tls encryption -> param3: false
  client, err := cacheClient.New("127.0.0.1", 8000, false, "", "")
  if err != nil {
    fmt.Println(err)
    return
  }

  // writing to connected cache to key "remote" with val "test1"
  client.AddKeyVal("remote", []byte("test1"))
  fmt.Println("Written val test1 to key remote")

  // requesting key value from connected cache from key "remote"
  fmt.Println("Read val from key remote: "+string(client.GetKeyVal("remote")))

  indexTestInstance(client)

  // starting testing routines
  // go concurrentTestInstanceA(client)
  // concurrentTestInstanceB(client)
}

func indexTestInstance(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    fmt.Println("index "+strconv.Itoa(i) + ": " + string(client.GetIndexVal(i)))
    time.Sleep(1 * time.Millisecond)
  }
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
