package main

import (
  "fmt"
  "remoteCacheToGo/cache"
  "strconv"
  "time"
)

func main() {
  // creating new cache database
  remoteCache := cache.New()

  remoteCache.AddValByKey("testKey", []byte("test1"))

  fmt.Println("Written val to testKey")

  // pulling data from cache "test" at key "testKey"
  res := remoteCache.GetValByKey("testKey")
  fmt.Println("Requestd key: " + string(res))

  // creating unencrypted network interfce for cache with name "remote"
  remoteCache.RemoteConnHandler(8000)
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
