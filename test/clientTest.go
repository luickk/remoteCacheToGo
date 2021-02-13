package main

import (
  "fmt"
  "strconv"
  // "time"
  "remoteCacheToGo/cacheClient"
)

func main() {
  fmt.Println("Client test")

  // creates new cacheClient struct and connects to remoteCache instance
  // no tls encryption -> param3: false
  client := cacheClient.New()
  if err := client.ConnectToCache("127.0.0.1", 8000, "", ""); err != nil {
    fmt.Println(err)
    return
  }

  // writing to connected cache to key "remote" with val "test1"
  if err := client.AddValByKey("remote", []byte("test1")); err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println("Written val test1 to key remote")

  // requesting key value from connected cache from key "remote"
  res, err := client.GetValByKey("remote")
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println("Read val from key remote: "+string(res))

  go concurrentWriteTest(client)

  // starting testing functions

  subscriptionTest(client)
  // concurrentGetTest(client)
}

func subscriptionTest(client cacheClient.RemoteCache) {
  sCh := client.Subscribe()
  for {
    select {
    case res := <-sCh:
      fmt.Println(res.Key +  ": " + string(res.Data))
    }
  }
}

func concurrentWriteTest(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    if err := client.AddValByKey("remote"+strconv.Itoa(i), []byte("remote"+strconv.Itoa(i))); err != nil {
      fmt.Println(err)
      break
    }
    // time.Sleep(1 * time.Millisecond)
  }
}

func concurrentGetTest(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    res, err := client.GetValByKey("remote"+strconv.Itoa(i))
    if err != nil {
      fmt.Println(err)
      break
    }
    fmt.Println("remote"+strconv.Itoa(i) + ": " + string(res))
    // time.Sleep(2 * time.Millisecond)
    }
}
