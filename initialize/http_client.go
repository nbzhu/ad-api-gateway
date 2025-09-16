package initialize

import (
	"github.com/nbzhu/ad-api-gateway/global"
	"github.com/nbzhu/ad-api-gateway/pkg"
	"time"
)

func InitHttpClient(timeout time.Duration) {
	global.Http = pkg.NewHttpClient(timeout)
}
