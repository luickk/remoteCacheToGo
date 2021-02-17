package main

import (
  "fmt"
  "time"
  "strconv"
  // "time"
  "remoteCacheToGo/cacheClient"
)

func main() {
  fmt.Println("Client test")
  errorStream := make(chan error)
  // creates new cacheClient struct and connects to remoteCache instance
  // no tls encryption -> param3: false
  client := cacheClient.New()
  go client.ConnectToCache("127.0.0.1", 8000, "", "", errorStream)

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

  go concurrentWriteTest(client, errorStream)

  // starting testing functions
  // go subscriptionTest(client)
  go concurrentGetTest(client)

  if err := <- errorStream; err != nil {
    fmt.Println(err)
    return
  }
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

func concurrentWriteTest(client cacheClient.RemoteCache, errorStream chan error) {
  i := 0
  for {
    i++
    if err := client.AddValByKey("remote"+strconv.Itoa(i), []byte("remote"+strconv.Itoa(i))); err != nil {
      errorStream <- err
      return
    }
    time.Sleep(1 * time.Millisecond)
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
    time.Sleep(2 * time.Millisecond)
    }
}
