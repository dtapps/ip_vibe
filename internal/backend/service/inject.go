package service

import (
	"sync"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_stat"

	"go.dtapp.net/library/contrib/http_log"
)

var (
	serviceDb     *ent.Client
	serviceDbOnce sync.Once
)

// 注入数据库客户端
func InjectDb(db *ent.Client) {
	serviceDbOnce.Do(func() {
		serviceDb = db
	})
}

var (
	serviceLogDb     *ent_log.Client
	serviceLogDbOnce sync.Once
)

// 注入日志数据库客户端
func InjectLogDb(logDb *ent_log.Client) {
	serviceLogDbOnce.Do(func() {
		serviceLogDb = logDb
	})
}

var (
	serviceStatDb     *ent_stat.Client
	serviceStatDbOnce sync.Once
)

// 注入统计数据库客户端
func InjectStatDb(statDb *ent_stat.Client) {
	serviceStatDbOnce.Do(func() {
		serviceStatDb = statDb
	})
}

var (
	serviceHttpTransport     *http_log.LoggingRoundTripper
	serviceHttpTransportOnce sync.Once
)

// 注入HTTP传输客户端
func InjectHttpTransport(httpTransport *http_log.LoggingRoundTripper) {
	serviceHttpTransportOnce.Do(func() {
		serviceHttpTransport = httpTransport
	})
}
