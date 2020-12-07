package main

import (
  "fmt"
  "strconv"
  "time"
  "remoteCacheToGo/cmd/cacheClient"
)


// optional, only requried if TLS is used
// !Cert ONLY for testing purposes!
const rootCert = `-----BEGIN CERTIFICATE-----
MIIFPzCCAyegAwIBAgIJAM2XHWKChMveMA0GCSqGSIb3DQEBBQUAMFIxCzAJBgNV
BAYTAkNOMQwwCgYDVQQIDANQRUsxETAPBgNVBAcMCEJlaSBKaW5nMQ8wDQYDVQQK
DAZWTXdhcmUxETAPBgNVBAMMCEhhcmJvckNBMB4XDTIwMTIwMzE5MjYyMFoXDTIx
MTIwMzE5MjYyMFowVzELMAkGA1UEBhMCQ04xDDAKBgNVBAgMA1BFSzERMA8GA1UE
BwwIQmVpIEppbmcxDzANBgNVBAoMBlZNd2FyZTEWMBQGA1UEAwwNSGFyYm9yTWFu
YWdlcjCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAPGt47nI62Y8Pe4f
tUbHb8fTgo2NKZyHsictuuNGRnqGlhWWu6NWCx/qT3D2Xm6pZUiCUWAp9A4WCHct
lIsFZeitFC1f8b4KIjSvQYqyiuRR4c2sMpvkg4cDgP9dvQgtIMCxFfB/NF+FmWNi
aqIRHf0XeEMK1yWgyrpcrz61zi2mKvU2rnQf+RaXo3iTnx6p0lOTQd3xx4gou+Uq
CkAJBqGunbWaPpkhIp1Dym7aj6mvZEpb3ojDPQ1v30mIsdMT0bwXgRPk/QweVkSQ
uA3p2yLbPB2u+/MWiRb2K6qpb1MfJ+R6wF1k8WQRRuH5PD8oNylEtaBU5ANNmQgS
z3G/43qY0lPdEiYXDoGfo/lDwjYfvBHaniLasl+s+yIzD9lvpkjV/hpQ5jXhG4ok
KEZEEBRuHjnwkSHRFl1keK22gf4TzV7UH6D1hlht8jnU90FmTljsbkYRtfdBRYYc
7hobbyBz8YestuRybiB2F90D6dqlvNatp4n6NYLjehsw0w2Ch5XtzOBJrEDf5kVx
SFKea+1+SZBUTtwGQ3bVW12fz0AY6NMiNtr+rDTjgpMyIBrEkTNJxR/tb3NV+uNq
M+97IDhIKKdgcQ1t+8wBW9LbhpGOSjmeKhBGmYdb451Jo03SnpJqTItbVvJGUX5z
qVC5l/pYtViObhUdhHwdbnD6V5Y1AgMBAAGjEzARMA8GA1UdEQQIMAaHBH8AAAEw
DQYJKoZIhvcNAQEFBQADggIBALDPSAP9yWZ2nMt6uK+w5w40ztNSTh7ywBninbYH
7kg0ZsgyRd52Doj20oqti6pgLkVOGn4iOEa+NN0XfCaPTiDLi9ha8mwA2VCl4OQ1
RYXROqQxbqTEz9RquyXVgM8OkOYCizbqWL48NOgxz9uOmcD39TT021lebG8xXBqY
Rn87l6SAH3NketQtcfFRLgtzjpP4RNle85//kS0zXJsjqUKYJIsWGTSq4sJTbS0F
vrr5HKQCtxIiktALdlTcxORkQxpSY0Efmmrb+C1LW3ih6RU+sCScF+5lrdvFkX2U
QxixG7XZDJcQI5CRYCze6k1Jl1waJb55/MKkT1eWQkpgb7cGiXpEMi+na9Lpz6Jd
lhx8N1alog2u9v0QY6tJX9Ic0Cc5mVrHP5Wik/qVn7n9CAGlQzZElJeugtrEDhqc
XkJj78W86O7iDb3MQmpYvSzPyTxmnBLloMBRxDPXr3ZgBHTxr+IgZ6taydUkYgRc
XMCYdcQsn+UaFyPArPbcCwTh6+r7hz5Nvrt4S7f/zG/w1tYLIxjx3b+UqAp1xzCx
TGbKqyGBKmFHRACE35Mme6Ekd1OK1YJmoOMPlOnf/5g+TlJSO16HOVpBnlLJ8ESV
JoGNvi2cyt37D0oUioYWJZPhr4ZHWfe9Hsbrq5lK/9qkNG7ozSHETUgxw4eb7q2w
X5/W
-----END CERTIFICATE-----
`

func main() {
  fmt.Println("Client test")

  // creates new cacheClient struct
  // params remote Cache IP, remote Cache port, wether connect with TLS encryption, root Cert for TLS encryption
  client, err := cacheClient.New("127.0.0.1", 8000, true, "test", rootCert)
  if err != nil {
    fmt.Println(err)
    return
  }

  // writing to connected cache to key "remote" with val "test1"
  client.AddKeyVal("remote", []byte("test1"))
  fmt.Println("Written val test1 to key remote")

  // requesting key value from connected cache from key "remote"
  fmt.Println("Read val from key remote: "+string(client.GetKeyVal("remote")))

  // starting testing routines
  go concurrentTestInstanceA(client)
  concurrentTestInstanceB(client)
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
