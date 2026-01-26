package utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

// ValidateCertAndKey 验证证书和私钥是否有效且匹配
func ValidateCertAndKey(certPEM, keyPEM string) error {
	// 1. 解析证书
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return fmt.Errorf("无效的证书格式")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("无效的证书内容")
	}

	// 2. 解析私钥
	keyBlock, _ := pem.Decode([]byte(keyPEM))
	if keyBlock == nil {
		return fmt.Errorf("无效的私钥格式")
	}

	var privKey crypto.PrivateKey
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		privKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	default:
		return fmt.Errorf("不支持的私钥类型")
	}
	if err != nil {
		return fmt.Errorf("无效的私钥内容")
	}

	// 3. 提取公钥
	var pubKey crypto.PublicKey
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		pubKey = k.Public()
	case *ecdsa.PrivateKey:
		pubKey = k.Public()
	default:
		return fmt.Errorf("不支持的私钥算法")
	}

	// 4. 比较公钥
	if !publicKeysEqual(cert.PublicKey, pubKey) {
		return fmt.Errorf("证书与私钥不匹配")
	}

	return nil
}

// ExtractDomainFromCert 从证书中提取主域名（CommonName）
func ExtractDomainFromCert(certPEM string) (string, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("无效的证书格式")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("解析证书失败: %w", err)
	}
	return cert.Subject.CommonName, nil
}

// publicKeysEqual 手动比较两个公钥是否相等
func publicKeysEqual(a, b crypto.PublicKey) bool {
	switch aPub := a.(type) {
	case *rsa.PublicKey:
		bPub, ok := b.(*rsa.PublicKey)
		if !ok {
			return false
		}
		return aPub.N.Cmp(bPub.N) == 0 && aPub.E == bPub.E

	case *ecdsa.PublicKey:
		bPub, ok := b.(*ecdsa.PublicKey)
		if !ok {
			return false
		}
		if aPub.Curve != bPub.Curve {
			return false
		}
		return aPub.X.Cmp(bPub.X) == 0 && aPub.Y.Cmp(bPub.Y) == 0

	default:
		return false
	}
}

// ParsedCertificate 从 PEM 证书中解析出的关键信息
type ParsedCertificate struct {
	Domain    string    // CommonName
	NotBefore time.Time // 生效时间
	NotAfter  time.Time // 过期时间
}

// ParseCertificate 从 PEM 格式证书中提取域名和有效期
func ParseCertificate(certPEM []byte) (*ParsedCertificate, error) {
	if len(certPEM) == 0 {
		return nil, fmt.Errorf("证书数据为空")
	}

	// 解析 PEM 数据
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("无效的证书格式")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	return &ParsedCertificate{
		Domain:    cert.Subject.CommonName,
		NotBefore: cert.NotBefore.UTC(),
		NotAfter:  cert.NotAfter.UTC(),
	}, nil
}
