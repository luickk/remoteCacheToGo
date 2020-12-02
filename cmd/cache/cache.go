package cache

import (
	"net"
	"fmt"
	"bufio"
	"bytes"
	"strconv"
	"crypto/tls"
  "remoteCacheToGo/internal/remoteCacheToGo"
)

type PushPullRequest struct {
	Key string

	ReturnPayload chan []byte
	Data []byte
}

type Cache struct {
	cacheMap map[string][]byte
	PushPullRequestCh chan *PushPullRequest
	CacheHandlerStarted bool
}

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 2048

// optional, only requried if TLS is used
// !Cert/ PKey ONLY for testing purposes!
const serverKey = `-----BEGIN EC PARAMETERS-----
BggqhkjOPQMBBw==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHg+g2unjA5BkDtXSN9ShN7kbPlbCcqcYdDu+QeV8XWuoAoGCCqGSM49
AwEHoUQDQgAEcZpodWh3SEs5Hh3rrEiu1LZOYSaNIWO34MgRxvqwz1FMpLxNlx0G
cSqrxhPubawptX5MSr02ft32kfOlYbaF5Q==
-----END EC PRIVATE KEY-----
`

const serverCert = `-----BEGIN CERTIFICATE-----
MIIB+TCCAZ+gAwIBAgIJAL05LKXo6PrrMAoGCCqGSM49BAMCMFkxCzAJBgNVBAYT
AkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRn
aXRzIFB0eSBMdGQxEjAQBgNVBAMMCWxvY2FsaG9zdDAeFw0xNTEyMDgxNDAxMTNa
Fw0yNTEyMDUxNDAxMTNaMFkxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0
YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxEjAQBgNVBAMM
CWxvY2FsaG9zdDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABHGaaHVod0hLOR4d
66xIrtS2TmEmjSFjt+DIEcb6sM9RTKS8TZcdBnEqq8YT7m2sKbV+TEq9Nn7d9pHz
pWG2heWjUDBOMB0GA1UdDgQWBBR0fqrecDJ44D/fiYJiOeBzfoqEijAfBgNVHSME
GDAWgBR0fqrecDJ44D/fiYJiOeBzfoqEijAMBgNVHRMEBTADAQH/MAoGCCqGSM49
BAMCA0gAMEUCIEKzVMF3JqjQjuM2rX7Rx8hancI5KJhwfeKu1xbyR7XaAiEA2UT7
1xOP035EcraRmWPe7tO0LpXgMxlh2VItpc2uc2w=
-----END CERTIFICATE-----
`

func (cache Cache) CacheHandler() {
	cache.CacheHandlerStarted = true
	for {
		select {
		case ppCacheOp := <-cache.PushPullRequestCh:
			if len(ppCacheOp.Data) <= 0 { // pull operation
				ppCacheOp.ReturnPayload <- cache.cacheMap[ppCacheOp.Key]
			} else if len(ppCacheOp.Data) > 0 { // push operation
				cache.cacheMap[ppCacheOp.Key] = ppCacheOp.Data
			}
		}
	}
}

func (cache Cache) RemoteConnHandler(port int) {
	l, err := net.Listen("tcp4", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go func(c net.Conn, cache Cache) {
			fmt.Printf("New client connected to %s \n", c.RemoteAddr().String())
			for {
				data := make([]byte, tcpConnBuffer)
				n, err := bufio.NewReader(c).Read(data)
				if err != nil {
					fmt.Println(err)
					return
				}
				data = data[:n]

				netDataSeperated := bytes.Split(data, []byte("\rnr"))
				if err != nil {
					fmt.Println(err)
					return
				}

				for _, data := range netDataSeperated {
					// fmt.Println(strconv.Itoa(len(data)) + ": " + string(data))
					if len(data) >= 1 {
							dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 3)
							if len(dataDelimSplitByte) >= 3 {
								key := string(dataDelimSplitByte[0])
								operation := string(dataDelimSplitByte[1])
								payload := dataDelimSplitByte[2]
								if operation == ">" { //pull
									c.Write(append(append([]byte(key+"->-"), cache.GetKeyVal(key)...), []byte("\rnr")...))
								} else if operation == "<" { // push
									cache.AddKeyVal(key, payload)
								}
							} else {
									fmt.Println("parsing error")
							}
						}
					}
				}
		}(c, cache)
	}
}

func (cache Cache) RemoteTlsConnHandler(port int) {
	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
		if err != nil {
			fmt.Println(err)
		}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	l, err := tls.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port), config)
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go func(c net.Conn, cache Cache) {
			fmt.Printf("New client connected to %s \n", c.RemoteAddr().String())
			for {
				data := make([]byte, tcpConnBuffer)
				n, err := bufio.NewReader(c).Read(data)
				if err != nil {
					fmt.Println(err)
					return
				}
				data = data[:n]

				netDataSeperated := bytes.Split(data, []byte("\rnr"))
				if err != nil {
					fmt.Println(err)
					return
				}

				for _, data := range netDataSeperated {
					// fmt.Println(strconv.Itoa(len(data)) + ": " + string(data))
					if len(data) >= 1 {
							dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 3)
							if len(dataDelimSplitByte) >= 3 {
								key := string(dataDelimSplitByte[0])
								operation := string(dataDelimSplitByte[1])
								payload := dataDelimSplitByte[2]
								if operation == ">" { //pull
									c.Write(append(append([]byte(key+"->-"), cache.GetKeyVal(key)...), []byte("\rnr")...))
								} else if operation == "<" { // push
									cache.AddKeyVal(key, payload)
								}
							} else {
									fmt.Println("parsing error")
							}
						}
					}
				}
		}(c, cache)
	}
}

func New() Cache {
  cache := Cache{make(map[string][]byte), make(chan *PushPullRequest), false}
	cache.CacheHandlerStarted = false
	go cache.CacheHandler()
  return cache
}

func (cache Cache) AddKeyVal(key string, val []byte) bool {
  if util.CharacterWhiteList(key) {
		request := new(PushPullRequest)
		request.Key = key
		request.Data = val

	  cache.PushPullRequestCh <- request

		request = nil
		return true
	}
	return false
}

func (cache Cache) GetKeyVal(key string) []byte {
  if util.CharacterWhiteList(key) {
		request := new(PushPullRequest)
		request.Key = key
		request.ReturnPayload = make(chan []byte)
	  cache.PushPullRequestCh <- request

		reply := false
		payload := []byte{}

		// wainting for request to be processed and retrieval of payload
		for !reply {
			select {
			case liveDataRes := <-request.ReturnPayload:
				payload = liveDataRes
				reply = true
				break
			}
		}
		request = nil
		return payload
	}
	return []byte{}
}
