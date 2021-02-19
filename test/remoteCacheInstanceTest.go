package main

import (
  "fmt"
  "strconv"
  "time"

  "remoteCacheToGo/cache"
)

func main() {
cert := `-----BEGIN CERTIFICATE-----
MIIFPzCCAyegAwIBAgIJAI8+kdLEcyiUMA0GCSqGSIb3DQEBBQUAMFIxCzAJBgNV
BAYTAkNOMQwwCgYDVQQIDANQRUsxETAPBgNVBAcMCEJlaSBKaW5nMQ8wDQYDVQQK
DAZWTXdhcmUxETAPBgNVBAMMCEhhcmJvckNBMB4XDTIxMDIxMzEwNDQxMloXDTIy
MDIxMzEwNDQxMlowVzELMAkGA1UEBhMCQ04xDDAKBgNVBAgMA1BFSzERMA8GA1UE
BwwIQmVpIEppbmcxDzANBgNVBAoMBlZNd2FyZTEWMBQGA1UEAwwNSGFyYm9yTWFu
YWdlcjCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAMDihRY9v+hBm+T5
VJl/9SIfocwgCDiM/oGgftrH/hzk8yDI2ztzFFLbsmpE75yPkntSy8nl8wulRUQQ
k6Ih+ZXEPySctyY0skdIw9UAfSaokFm7lfKDSifN0pnr/MCTWDyPrC/Srzt2xDlg
MaBAoh95Q5QahSYpP3k6s0uS0nC8HU7EmqY7xgxBLtlOO8uvewfC5xs0KTA69sxJ
JkBdkUCobPcDFE4r7d4mWiOQf50/+j0TU/fwcytQ7Gk+bhXKgKYeiaZQLTcmiBwe
7+Llvig8gPgj4ftEwWm3gMGYWurd+os2YoMUtFNp8Hib9bV35tU1qvwc8gfS+3iv
oeTlr6qrY3s9i9+qvwsZ/nQHpzrrqANXquT1u2BFO1qwTTF9hQyFm+mnjuhFBISx
dx/OWkWHEnuNCnzcdlBsE2sWlEMOprdbCJFxex1uBlyo9cxbxhrfZiZGCOremLwv
UHwl5C6dIZhNfuizpOwClHXxNLMgNPhRruX2aPctmM8zGwcQiliKvsr7YuIlnYXe
BW9yTbFpt5Nh1liXk+yQJR+wOp/ofajXnVgkvbzjy7oReAmVCXYmLEtUFpWhItFd
dwr50sH3VIO8s2JFuMbCv/4KzHukvgsN5h0dMRvXwNjaA5YkLtDwFkLhrWOEX3Hj
Gr+iK/1aZv4LPKAEKu0cwz4KrZnXAgMBAAGjEzARMA8GA1UdEQQIMAaHBH8AAAEw
DQYJKoZIhvcNAQEFBQADggIBACkgBWK8CclDfbhyFQFSe0rwh1fmmVljn/1dQaJf
OpGgbfrl15g/05VK9tIV30sUYcDoYS8/ZStGuHl4wX3Y+lyTzmZsjPSTu/0wc6T5
PNHug3z5xi+N5xievhIllKJPywDpWKM7X+t9fPSeVapKbpUNr+lCH4hdOYLaVXjO
EanU/3lSEaDflNtywLrZ5xg8KHJ2AWfOCPoJmuLdkbcXFLSv86empYIBY8lEYJCb
svk6AmeJSo42oHaJ7equFzmBYOsEtP3cm3mos/9ui06mU+KQW0xdrwGGEyTisq6e
TplqiJ2iGLkh0qupuvRFYkKu08RAfkCAZHKKWUJA2QgNms/XM+b8ttGt+S5ydrUh
MB0VW1SNirDfAvCYXzxgWLbRk6CInThxJ+q3QDFqk843iOpd/YNfGD6jW9hO6aJK
DaUQbwsWWg4MAwvd8bsfJ9WwB/zWiQfwLLrR8oxAyt3PU0qI+5xaq9lp2Pl4naA6
lGoag61ek8G6oKIAw6XXUy+UKJ81wIA7qajCpOrXFFW3QgWhFU3wwyixFNFWUj/H
tDmsJUs9P/27U7uS2cQy+u1VTvJFRKAfpLG38AZeAA+ttoT/N8IOTBE7NhKtTt8j
ZmMKcSu5NdUBt5Iz10kM7AM+eyF77obzdN3ixB7XrpzflnF4yGjp2LlUhkTCpBCk
dx2R
-----END CERTIFICATE-----`

key := `-----BEGIN PRIVATE KEY-----
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
q+G17vXnLprHmbcZabzObzLD0WZdBTI=
-----END PRIVATE KEY-----`
  errorStream := make(chan error)
  // creating new cache database
  remoteCache := cache.New(errorStream)

  remoteCache.AddValByKey("testKey", []byte("test1"))

  fmt.Println("Written val to testKey")

  // pulling data from cache "test" at key "testKey"
  res := remoteCache.GetValByKey("testKey")
  fmt.Println("Requestd key: " + string(res))

  // creating unencrypted network interfce for cache with name "remote"
  go remoteCache.RemoteTlsConnHandler(8000, "127.0.0.1", "test", cert, key, errorStream)

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
