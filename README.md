# Remote Cache To Go

This project creates a remote-database and an embedded-cache storage hybrid which enables live data synchronisation as well as encrypted and authorised data exchange.
By maintaining go's idiomatic routine/ channel concept, concurrency is built into the program by design. No packages are used to reduce dependencies and complexity.

## Concept

All cache objects are referenced in the cache database but can also be implemented directly. Every cache can be made remotely accessible independently or just be used in a local scope. If made remotely accessible, a port has to be provided (in case TLS is enabled a public/ private key pair has to be generated passed as params). With the client part of the package it can then be easily connected to the remote cache and treated just like a local object.

## Specifications

The cacheMap is a simple map type object which is held in cache which enables fast and temporary data storage, exchange or buffering. The remote data transmission is done via. TCP sockets and the information is encoded in a null terminated json string. As a general delimiter to differentiate between the json strings a null byte is used.

## Features

- concurrency by go idiomacy
- local key/ val cache
the cache can only be used in a local program context
- remote key/ val cache  <br>
the cache can be made accessible via tcp conn. either encrypted(tls) or unencrypted
- key/val cache can be used as queue <br>
Elements can be requested by the order(index or count) they have been pushed to the cache


## Example


Server:
``` go

// creating new cache database
remoteCache := cache.New()

remoteCache.AddValByKey("testKey", []byte("test1"))

fmt.Println("Written val to testKey")

// pulling data from cache "test" at key "testKey"
res := remoteCache.GetValByKey("testKey")
fmt.Println("Requestd key: " + string(res))

// creating unencrypted network interfce for cache with name "remote"
remoteCache.RemoteConnHandler(8000)

// creating encrypted network interface for cache with name "remote" and the password hash "test" and enabled dosProtection
// serverCert & Key are passed hardcoded only for testing purposes
remoteCache.RemoteTlsConnHandler(8001, "test", true, serverCert, serverKey)

}

```

Client:
``` go
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

```

#### Comm Protocol

Communication happens by json marshalling a predefined struct (SPushPullReq struct) and unmarshalling it on the receiver side.
The marsahlled byte arrays are exchanged via. a field length framed tcp connection.
`<8 byte 64 bit unsigned integer containing the data field length><data>`

The marshalled struct:
``` go
type SPushPullReq struct {
	Type string
	Name string
	Id int
	Operation string
	Payload []byte
}
```

## Sketch of routine layout

![sketch.png](media/sketch.png)
