package main

import (
	"github.com/nbzhu/ad-api-gateway/initialize"
	"time"
)

func main() {
	initialize.InitHttpClient(10 * time.Second)
	initialize.InitServer(50051)
}
