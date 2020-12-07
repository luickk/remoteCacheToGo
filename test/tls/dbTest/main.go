package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheDb"
  "strconv"
  "time"
)


// optional, only requried if TLS is used
// !Cert/ PKey ONLY for testing purposes!
const serverKey = `-----BEGIN PRIVATE KEY-----
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
yHMvBkQ+XqcTJJqxxDJhPnVFN3uONQ==
-----END PRIVATE KEY-----
`

const serverCert = `-----BEGIN CERTIFICATE-----
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
	fmt.Println("DB test")

  // creating new cache database
  cDb := cacheDb.New()

  // creating new caches in database
  cDb.NewCache("test")
  cDb.NewCache("remote")

  // adding new entry to cache "test" at key "testkey" with val "test1"
  if cDb.AddEntryToCache("test", "testKey", []byte("test1")) {
    fmt.Println("Written val to testKey")
  }

  // pulling data from cache "test" at key "testKey"
  fmt.Println("Requestd key: " + string(cDb.GetEntryFromCache("test", "testKey")))

  // creating encrypted network interface for cache with name "remote" and the password hash "test" and enabled dosProtection
  // serverCert & Key are passed hardcoded only for testing purposes
  cDb.Db["remote"].RemoteTlsConnHandler(8000, "test", true, serverCert, serverKey)


  // runing test instances
  go concurrentTestInstanceA(cDb)
  concurrentTestInstanceB(cDb)
}

func concurrentTestInstanceA(cDb cacheDb.CacheDb) {
  i := 0
  for {
    i++
    cDb.AddEntryToCache("test", "test"+strconv.Itoa(i), []byte("test"+strconv.Itoa(i)))
  }
}

func concurrentTestInstanceB(cDb cacheDb.CacheDb) {
  i := 0
  for {
    i++
    fmt.Println("test"+strconv.Itoa(i) + ": " + string(cDb.GetEntryFromCache("test", "test"+strconv.Itoa(i))))
    fmt.Println("remote: "+string(cDb.GetEntryFromCache("remote", "remote")))
    time.Sleep(10 * time.Millisecond)
    }
}
