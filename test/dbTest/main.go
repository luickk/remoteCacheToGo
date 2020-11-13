package main

import (
  "fmt"
  "remoteCacheToGo/cmd/cacheDb"
)

func main() {
	fmt.Println("DB test")
  dbs := cacheDb.New()
}
