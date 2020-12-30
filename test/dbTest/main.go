package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheDb"
  "strconv"
  "time"
)

func main() {
	fmt.Println("DB test")

  // creating new cache database
  cDb := cacheDb.New()

  // creating new caches in database
  cDb.NewCache("test")
  cDb.NewCache("remote")

  if err := cDb.Db["test"].AddValByKey("testKey", []byte("test1")); err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println("Written val to testKey")

  // pulling data from cache "test" at key "testKey"
  res, err := cDb.Db["test"].GetValByKey("testKey")
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println("Requestd key: " + string(res))

  // creating unencrypted network interfce for cache with name "remote"
  cDb.Db["remote"].RemoteConnHandler(8000)

  // runing test instances
  // go concurrentTestInstanceA(cDb)
  // concurrentTestInstanceB(cDb)
}

func concurrentTestInstanceA(cDb cacheDb.CacheDb) {
  i := 0
  for {
    i++
    cDb.Db["test"].AddValByKey("test"+strconv.Itoa(i), []byte("test"+strconv.Itoa(i)))
  }
}

func concurrentTestInstanceB(cDb cacheDb.CacheDb) {
  i := 0
  for {
    i++
    res, err := cDb.Db["test"].GetValByKey("test"+strconv.Itoa(i))
    if err != nil {
      fmt.Println(err)
    }
    fmt.Println("test"+strconv.Itoa(i) + ": " + string(res))
    res, err = cDb.Db["remote"].GetValByKey("remote"+strconv.Itoa(i))
    if err != nil {
      fmt.Println(err)
    }
    fmt.Println("remote: "+string(res))
    time.Sleep(10 * time.Millisecond)
    }
}
