package main

import (
  "fmt"
  "strconv"
  "time"

  "remoteCacheToGo/cache"
)

func main() {
  errorStream := make(chan error)
  // creating new cache database
  remoteCache := cache.New(errorStream)

  remoteCache.AddValByKey("testKey", []byte("test1"))

  fmt.Println("Written val to testKey")

  // pulling data from cache "test" at key "testKey"
  res := remoteCache.GetValByKey("testKey")
  fmt.Println("Requestd key: " + string(res))

  // go remoteCache.RemoteTlsConnHandler(8000, "127.0.0.1", "", "", "", errorStream)

  // creating unencrypted network interfce for cache with name "remote"
  go remoteCache.RemoteConnHandler("127.0.0.1", 8000, errorStream)

  for {
    if err := <- errorStream; err != nil {
      fmt.Println(err)
    }
  }
}

func concurrentTestInstanceA(remoteCache cache.Cache) {
  i := 0
  for {
    i++
    remoteCache.AddValByKey("test"+strconv.Itoa(i), []byte("test"+strconv.Itoa(i)))
  }
}

func concurrentTestInstanceB(remoteCache cache.Cache) {
  i := 0
  for {
    i++
    res := remoteCache.GetValByKey("test"+strconv.Itoa(i))
    fmt.Println("test"+strconv.Itoa(i) + ": " + string(res))
    res = remoteCache.GetValByKey("remote"+strconv.Itoa(i))
    fmt.Println("remote: "+string(res))
    time.Sleep(10 * time.Millisecond)
    }
}
