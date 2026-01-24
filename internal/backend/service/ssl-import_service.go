package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/sslimport"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"
)

// SSLImportService SSL 记录服务接口
type SSLImportService interface {
	BuildResponse(ctx context.Context, item *ent.SSLImport) (*types.GetSSLImportResponse, error)
	ImportSSLImport(ctx context.Context, request *types.ImportSSLImportRequest) error
	CreateSSLImport(ctx context.Context, request *types.CreateSSLImportRequest) error
	DeleteSSLImport(ctx context.Context, id int64) error
	BatchDeleteSSLImport(ctx context.Context, ids []int64) error
}

// SSLImportServiceImpl SSL 记录服务实现
type SSLImportServiceImpl struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器
	debug  bool            // 调试模式
}

// NewSSLImportService 创建 SSL 记录服务实例
func NewSSLImportService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) SSLImportService {
	return &SSLImportServiceImpl{
		db:     db,
		logDb:  logDb,
		logger: logger,
		debug:  sys_env.GvaEnv.DEBUG,
	}
}

// BuildResponse 将实体转换为响应
func (s *SSLImportServiceImpl) BuildResponse(ctx context.Context, item *ent.SSLImport) (*types.GetSSLImportResponse, error) {
	resp := &types.GetSSLImportResponse{
		ID:          item.ID,
		Domain:      item.Domain,
		Certificate: item.Certificate,
		PrivateKey:  item.PrivateKey,
		NotBefore:   item.NotBefore,
		NotAfter:    item.NotAfter,
	}

	return resp, nil
}

// ImportSSLImport 导入 SSL 导入记录
func (s *SSLImportServiceImpl) ImportSSLImport(ctx context.Context, request *types.ImportSSLImportRequest) error {

	// 解析证书
	parsedSsl, err := utils.ParseCertificate(request.Certificate)
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	// 创建SSL导入记录
	err = s.db.SSLImport.Create().
		SetSource(sslimport.SourceWebhook).
		SetDomain(parsedSsl.Domain).
		SetCertificate(request.Certificate).
		SetPrivateKey(request.PrivateKey).
		SetNotBefore(parsedSsl.NotBefore).
		SetNotAfter(parsedSsl.NotAfter).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建SSL导入记录失败: %w", err)
	}

	return nil
}

// CreateSSLImport 创建 SSL 导入记录
func (s *SSLImportServiceImpl) CreateSSLImport(ctx context.Context, request *types.CreateSSLImportRequest) error {

	// 解析证书
	parsedSsl, err := utils.ParseCertificate(request.Certificate)
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	// 验证证书与私钥是否匹配
	err = utils.ValidateCertAndKey(request.Certificate, request.PrivateKey)
	if err != nil {
		return fmt.Errorf("验证证书与私钥匹配失败: %w", err)
	}

	// 创建SSL记录
	err = s.db.SSLImport.Create().
		SetSource(sslimport.SourceManual).
		SetDomain(request.Domain).
		SetCertificate(request.Certificate).
		SetPrivateKey(request.PrivateKey).
		SetNotBefore(parsedSsl.NotBefore).
		SetNotAfter(parsedSsl.NotAfter).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建SSL导入记录失败: %w", err)
	}

	return nil
}

// DeleteSSLImport 删除 CDN 导入记录
func (s *SSLImportServiceImpl) DeleteSSLImport(ctx context.Context, id int64) error {
	// 查询SSL记录
	firstSSLImport, err := s.db.SSLImport.Query().
		Where(sslimport.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询SSL导入记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询SSL导入记录失败: %w", err)
	}

	// 删除SSL导入记录
	err = s.db.SSLImport.DeleteOne(firstSSLImport).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除SSL导入记录失败: %w", err)
	}

	return nil
}

// BatchDeleteSSLImport 批量删除 SSL 导入记录
func (s *SSLImportServiceImpl) BatchDeleteSSLImport(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询SSL记录
	allSSLImport, err := tx.SSLImport.Query().
		Where(sslimport.IDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询SSL记录失败: %w", err)
	}
	if len(allSSLImport) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询SSL记录失败: 数量 %d 与请求数量 %d 不匹配", len(allSSLImport), len(ids))
	}

	// 删除SSL记录
	_, err = tx.SSLImport.Delete().
		Where(sslimport.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除SSL记录失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}
