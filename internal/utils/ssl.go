package utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"reflect"
	"strings"
	"time"
	"vibe_ip/internal/types"
)

// ValidateCertAndKey 验证证书和私钥是否有效且匹配
func ValidateCertAndKey(certPEM []byte, keyPEM []byte) error {
	if len(certPEM) == 0 {
		return fmt.Errorf("证书数据为空")
	}
	if len(keyPEM) == 0 {
		return fmt.Errorf("私钥数据为空")
	}

	// 解析证书数据
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return fmt.Errorf("无效的证书格式")
	}

	// 解析证书内容
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("无效的证书内容")
	}

	// 解析私钥数据
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("无效的私钥格式")
	}

	// 解析私钥
	var privKey crypto.PrivateKey
	switch keyBlock.Type {
	case "PRIVATE KEY": // PKCS8 格式
		privKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	case "RSA PRIVATE KEY": // PKCS1
		privKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	default:
		return fmt.Errorf("不支持的私钥类型")
	}
	if err != nil {
		return fmt.Errorf("无效的私钥内容")
	}

	// 提取公钥
	var pubKey crypto.PublicKey
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		pubKey = k.Public()
	case *ecdsa.PrivateKey:
		pubKey = k.Public()
	default:
		return fmt.Errorf("不支持的私钥算法")
	}

	// 比较公钥
	if !publicKeysEqual(cert.PublicKey, pubKey) {
		return fmt.Errorf("证书与私钥不匹配")
	}

	return nil
}

// publicKeysEqual 手动比较两个公钥是否相等
func publicKeysEqual(a, b crypto.PublicKey) bool {
	// 判断类型是否一致
	if a == nil || b == nil || reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	switch aPub := a.(type) {
	case *rsa.PublicKey:
		if bPub, ok := b.(*rsa.PublicKey); ok {
			return aPub.E == bPub.E && aPub.N.Cmp(bPub.N) == 0
		}
		return false
	case *ecdsa.PublicKey:
		return aPub.Equal(b)
	default:
		return false
	}
}

// ParseCertificate 从 PEM 格式证书中提取域名和有效期
func ParseCertificate(certPEM []byte) (*types.UtilsParsedCertificate, error) {
	if len(certPEM) == 0 {
		return nil, fmt.Errorf("证书数据为空")
	}

	// 解析证书数据
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("无效的证书格式")
	}

	// 解析证书内容
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	return &types.UtilsParsedCertificate{
		Domain:    cert.Subject.CommonName,
		NotBefore: cert.NotBefore.UTC(),
		NotAfter:  cert.NotAfter.UTC(),
	}, nil
}

// ParseCertificate 从 PEM 格式证书中提取全部信息
func ParseCertificateFull(certPEM []byte) (*types.UtilsParsedCertificateFull, error) {
	if len(certPEM) == 0 {
		return nil, fmt.Errorf("证书数据为空")
	}

	// 解析证书数据
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("无效的证书格式")
	}

	// 解析证书内容
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 获取当前时间 (UTC)
	now := time.Now().UTC()

	// 判断是否包含通配符
	hasWildcard := false
	for _, name := range cert.DNSNames {
		if strings.Contains(name, "*") {
			hasWildcard = true
			break
		}
	}

	// 获取密钥长度
	keyLen := 0
	isECC := false
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		keyLen = pub.N.BitLen()
	case *ecdsa.PublicKey:
		isECC = true
		keyLen = pub.Curve.Params().BitSize
	}

	// 计算百分比
	total := cert.NotAfter.Sub(cert.NotBefore).Seconds()
	elapsed := now.Sub(cert.NotBefore).Seconds()
	percent := int((elapsed / total) * 100)
	if percent > 100 {
		percent = 100
	} else if percent < 0 {
		percent = 0
	}

	return &types.UtilsParsedCertificateFull{
		// 身份信息
		Subject:      cert.Subject.CommonName,
		Issuer:       cert.Issuer.CommonName,
		DNSNames:     cert.DNSNames,
		SerialNumber: cert.SerialNumber.String(),
		// 算法与安全
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		KeyLength:          keyLen,
		Version:            cert.Version,
		// 时间与寿命
		NotBefore:     cert.NotBefore,
		NotAfter:      cert.NotAfter,
		ExpiryDays:    int(cert.NotAfter.Sub(now).Hours() / 24),
		ExpiryPercent: percent,
		// 状态标志 (用于 UI 逻辑)
		IsExpired:  now.After(cert.NotAfter),
		IsWildcard: hasWildcard,
		IsECC:      isECC,
	}, nil
}
