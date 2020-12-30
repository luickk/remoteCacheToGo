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

  // indexTestInstance(client)
  // countTestInstance(client)
  // indexKeyTestInstance(client)

  // starting testing routines
  go concurrentTestInstanceA(client)
  concurrentTestInstanceB(client)
}

func indexTestInstance(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    fmt.Println("index "+strconv.Itoa(i) + ": " + string(client.GetValByIndex(i)))
    time.Sleep(1 * time.Millisecond)
  }
}

func indexKeyTestInstance(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    fmt.Println("index(key) "+strconv.Itoa(i) + ": " + string(client.GetKeyByIndex(i)))
    time.Sleep(1 * time.Millisecond)
  }
}

func countTestInstance(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    count, err := client.GetCountByIndex(i)
    if err != nil {
      fmt.Println(err)
    }
    fmt.Println("count(reversed Index) "+strconv.Itoa(i) + ": " + strconv.Itoa(count))
    time.Sleep(1 * time.Millisecond)
  }
}


func concurrentTestInstanceA(client cacheClient.RemoteCache) {
  i := 0
  for {
    i++
    if err := client.AddValByKey("remote"+strconv.Itoa(i), []byte("remote"+strconv.Itoa(i))); err != nil {
      fmt.Println(err)
      break
    }
    time.Sleep(1 * time.Millisecond)
  }
}

func concurrentTestInstanceB(client cacheClient.RemoteCache) {
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
