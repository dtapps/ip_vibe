package service

import (
	"sync"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_stat"
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
