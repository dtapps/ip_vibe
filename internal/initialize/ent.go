package initialize

import (
	"context"
	"database/sql"
	"log"
	"vibe_ip/ent"
	"vibe_ip/internal/global"

	"entgo.io/ent/dialect"
	entSql "entgo.io/ent/dialect/sql"
	"modernc.org/sqlite"
)

func init() {
	sql.Register(dialect.SQLite, &sqlite.Driver{})
}

// InitEnt 初始化 ent
func InitEnt(ctx context.Context) {
	var err error

	// SQLite DSN，启用 WAL + 外键 + 超时
	dsn := "file:data/ent.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout=5000"

	// 1. 先打开原生 *sql.DB
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("[初始化 ent] 打开数据库失败:%v", err)
	}

	// 2. 设置连接池（防止并发写冲突）
	sqlDB.SetMaxOpenConns(1) // 串行写，防止 database is locked
	sqlDB.SetMaxIdleConns(1)

	// 3. 使用 entsql.OpenDB 包装为 ent 驱动
	driver := entSql.OpenDB(dialect.SQLite, sqlDB)

	// 3. 让 Ent 使用这个 DB  dialect.SQLite,
	global.GvaDb = ent.NewClient(ent.Driver(driver))

	// 4. 开启调试模式（仅在开发环境）
	if global.GvaEnv.DEBUG {
		global.GvaDb = global.GvaDb.Debug()
	}

	// 5. 自动迁移
	if err := global.GvaDb.Schema.Create(ctx); err != nil {
		log.Fatalf("[初始化 ent] 自动迁移失败:%v", err)
	}
}

// CloseEnt 关闭 ent
func CloseEnt() {
	if global.GvaDb != nil {
		if err := global.GvaDb.Close(); err != nil {
			log.Fatalf("[关闭 ent] :%v", err)
		}
	}
}
