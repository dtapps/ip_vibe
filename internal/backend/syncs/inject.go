package syncs

import (
	"sync"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_stat"

	"go.dtapp.net/library/contrib/http_log"
)

var (
	syncDb     *ent.Client
	syncDbOnce sync.Once
)

// 注入数据库客户端
func InjectDb(db *ent.Client) {
	syncDbOnce.Do(func() {
		syncDb = db
	})
}

var (
	syncLogDb     *ent_log.Client
	syncLogDbOnce sync.Once
)

// 注入日志数据库客户端
func InjectLogDb(logDb *ent_log.Client) {
	syncLogDbOnce.Do(func() {
		syncLogDb = logDb
	})
}

var (
	syncStatDb     *ent_stat.Client
	syncStatDbOnce sync.Once
)

// 注入统计数据库客户端
func InjectStatDb(statDb *ent_stat.Client) {
	syncStatDbOnce.Do(func() {
		syncStatDb = statDb
	})
}

var (
	syncHttpTransport     *http_log.LoggingRoundTripper
	syncHttpTransportOnce sync.Once
)

// 注入HTTP传输客户端
func InjectHttpTransport(httpTransport *http_log.LoggingRoundTripper) {
	syncHttpTransportOnce.Do(func() {
		syncHttpTransport = httpTransport
	})
}
