package initialize

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"vibe_ip/internal/ent"
	_ "vibe_ip/internal/ent/runtime"
	"vibe_ip/internal/ent_log"
	_ "vibe_ip/internal/ent_log/runtime"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/ent_slog"

	"entgo.io/ent/dialect"
	entSql "entgo.io/ent/dialect/sql"
	"modernc.org/sqlite"
)

func init() {
	sql.Register(dialect.SQLite, &sqlite.Driver{})
}

// InitEntDb 初始化 业务数据库
func InitEntDb(ctx context.Context, logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	var err error

	// SQLite DSN，启用 WAL + 外键 + 超时
	dsn := "file:data/ent.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout=5000"

	// 1. 先打开原生 *sql.DB
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		slog.ErrorContext(ctx, "[初始化 业务数据库] 打开数据库失败",
			slog.Any("err", err),
		)
		os.Exit(1)
	}

	// 2. 设置连接池（防止并发写冲突）
	sqlDB.SetMaxOpenConns(1) // 串行写，防止 database is locked
	sqlDB.SetMaxIdleConns(1)

	// 3. 使用 entsql.OpenDB 包装为 ent 驱动
	driver := entSql.OpenDB(dialect.SQLite, sqlDB)

	// 3. 让 Ent 使用这个 DB  dialect.SQLite,
	opts := make([]ent.Option, 0)
	opts = append(
		opts,
		ent.Driver(driver),
	)
	opts = append(
		opts,
		ent.Log(ent_slog.NewSlogLogger(logger, "ent").Log),
	)
	if global.GvaEnv.DEBUG {
		opts = append(
			opts,
			ent.Debug(),
		)
	}
	global.GvaEntDb = ent.NewClient(opts...)

	// 4. 自动迁移
	if err := global.GvaEntDb.Schema.Create(ctx); err != nil {
		slog.ErrorContext(ctx, "[初始化 业务数据库] 自动迁移失败",
			slog.Any("err", err),
		)
		os.Exit(1)
	}
}

// CloseEntDb 关闭 业务数据库
func CloseEntDb() {
	if global.GvaEntDb != nil {
		if err := global.GvaEntDb.Close(); err != nil {
			slog.Error("[关闭 业务数据库] 关闭错误",
				slog.Any("err", err),
			)
			os.Exit(1)
		}
	}
}

// InitEntLogDb 初始化 日志数据库
func InitEntLogDb(ctx context.Context, logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	var err error

	// SQLite DSN，启用 WAL + 外键 + 超时
	dsn := "file:data/ent_log.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout=5000"

	// 1. 先打开原生 *sql.DB
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		slog.ErrorContext(ctx, "[初始化 日志数据库] 打开数据库失败",
			slog.Any("err", err),
		)
		os.Exit(1)
	}

	// 2. 设置连接池（防止并发写冲突）
	sqlDB.SetMaxOpenConns(1) // 串行写，防止 database is locked
	sqlDB.SetMaxIdleConns(1)

	// 3. 使用 entsql.OpenDB 包装为 ent 驱动
	driver := entSql.OpenDB(dialect.SQLite, sqlDB)

	// 3. 让 Ent 使用这个 DB  dialect.SQLite,
	opts := make([]ent_log.Option, 0)
	opts = append(
		opts,
		ent_log.Driver(driver),
	)
	opts = append(
		opts,
		ent_log.Log(ent_slog.NewSlogLogger(logger, "ent_log").Log),
	)
	if global.GvaEnv.DEBUG {
		opts = append(
			opts,
			ent_log.Debug(),
		)
	}
	global.GvaEntLogDb = ent_log.NewClient(opts...)

	// 4. 自动迁移
	if err := global.GvaEntLogDb.Schema.Create(ctx); err != nil {
		slog.ErrorContext(ctx, "[初始化 日志数据库] 自动迁移失败",
			slog.Any("err", err),
		)
		os.Exit(1)
	}
}

// CloseEntLogDb 关闭 业务数据库
func CloseEntLogDb() {
	if global.GvaEntLogDb != nil {
		if err := global.GvaEntLogDb.Close(); err != nil {
			slog.Error("[关闭 日志数据库] 关闭错误",
				slog.Any("err", err),
			)
			os.Exit(1)
		}
	}
}
