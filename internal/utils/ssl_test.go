package utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

// testKey 封装一个测试用私钥：既有 crypto.Signer 用于签发自签证书，
// 也有对应格式的 PEM（用于覆盖 PKCS8/PKCS1/SEC1 三种解析分支）。
type testKey struct {
	signer  crypto.Signer
	keyPEM  []byte
	keyType string
}

// newRSAKey 生成 RSA 私钥，blockType 决定 PEM 块类型：
// "PRIVATE KEY"(PKCS8) 或 "RSA PRIVATE KEY"(PKCS1)。
func newRSAKey(t *testing.T, blockType string) testKey {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成 RSA 私钥失败: %v", err)
	}

	var der []byte
	switch blockType {
	case "PRIVATE KEY":
		der, err = x509.MarshalPKCS8PrivateKey(priv)
	case "RSA PRIVATE KEY":
		der = x509.MarshalPKCS1PrivateKey(priv)
	default:
		t.Fatalf("不支持的测试私钥块类型: %s", blockType)
	}
	if err != nil {
		t.Fatalf("序列化 RSA 私钥失败: %v", err)
	}

	return testKey{signer: priv, keyPEM: pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}), keyType: blockType}
}

// newECKey 生成 P-256 私钥，blockType 取 "PRIVATE KEY"(PKCS8) 或 "EC PRIVATE KEY"(SEC1)。
func newECKey(t *testing.T, blockType string) testKey {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成 ECDSA 私钥失败: %v", err)
	}

	var der []byte
	switch blockType {
	case "PRIVATE KEY":
		der, err = x509.MarshalPKCS8PrivateKey(priv)
	case "EC PRIVATE KEY":
		der, err = x509.MarshalECPrivateKey(priv)
	default:
		t.Fatalf("不支持的测试私钥块类型: %s", blockType)
	}
	if err != nil {
		t.Fatalf("序列化 ECDSA 私钥失败: %v", err)
	}

	return testKey{signer: priv, keyPEM: pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}), keyType: blockType}
}

// newSelfSignedCert 用给定私钥签发一张自签证书，返回 PEM 编码。
func newSelfSignedCert(t *testing.T, key testKey, tmpl *x509.Certificate) []byte {
	t.Helper()

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.signer.Public(), key.signer)
	if err != nil {
		t.Fatalf("签发自签证书失败: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

// certTemplate 构造证书模板：CN、SAN、序列号与有效期由调用方指定。
func certTemplate(cn string, serial int64, dnsNames []string, notBefore, notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{CommonName: cn},
		Issuer:                pkix.Name{CommonName: cn},
		DNSNames:              dnsNames,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
	}
}

// reencodeKey 把同一个私钥重新编码为指定 PEM 块类型，
// 用于验证「证书与私钥匹配」时不依赖密钥本身是否相同。
func reencodeKey(t *testing.T, signer crypto.Signer, blockType string) []byte {
	t.Helper()

	var (
		der []byte
		err error
	)
	switch blockType {
	case "PRIVATE KEY":
		der, err = x509.MarshalPKCS8PrivateKey(signer)
	case "RSA PRIVATE KEY":
		rsaPriv, ok := signer.(*rsa.PrivateKey)
		if !ok {
			t.Fatalf("RSA PRIVATE KEY 需要 RSA 私钥")
		}
		der = x509.MarshalPKCS1PrivateKey(rsaPriv)
	case "EC PRIVATE KEY":
		ecPriv, ok := signer.(*ecdsa.PrivateKey)
		if !ok {
			t.Fatalf("EC PRIVATE KEY 需要 ECDSA 私钥")
		}
		der, err = x509.MarshalECPrivateKey(ecPriv)
	default:
		t.Fatalf("不支持的测试私钥块类型: %s", blockType)
	}
	if err != nil {
		t.Fatalf("序列化私钥失败: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
}

func TestValidateCertAndKey(t *testing.T) {
	rsaKey := newRSAKey(t, "PRIVATE KEY")
	rsaOther := newRSAKey(t, "PRIVATE KEY")
	ecKey := newECKey(t, "PRIVATE KEY")
	ecOther := newECKey(t, "PRIVATE KEY")

	rsaPKCS8PEM := reencodeKey(t, rsaKey.signer, "PRIVATE KEY")
	rsaPKCS1PEM := reencodeKey(t, rsaKey.signer, "RSA PRIVATE KEY")
	ecPKCS8PEM := reencodeKey(t, ecKey.signer, "PRIVATE KEY")
	ecSEC1PEM := reencodeKey(t, ecKey.signer, "EC PRIVATE KEY")

	now := time.Now()
	tmpl := certTemplate("a.example.com", 1, []string{"a.example.com"}, now.Add(-time.Hour), now.Add(24*time.Hour))
	rsaCert := newSelfSignedCert(t, rsaKey, tmpl)
	ecCert := newSelfSignedCert(t, ecKey, tmpl)
	otherRSACert := newSelfSignedCert(t, rsaOther, tmpl)
	otherECCert := newSelfSignedCert(t, ecOther, tmpl)

	// PEM 块类型正确但内容是非证书字节，触发「无效的证书内容」
	invalidCertContent := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not a cert")})
	// PEM 块类型正确但内容不是合法 PKCS8 私钥，触发「无效的私钥内容」
	invalidKeyContent := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not a key")})

	// ed25519 私钥可被 PKCS8 解析，但不在函数支持的算法范围内
	_, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("生成 ed25519 私钥失败: %v", err)
	}
	edDER, err := x509.MarshalPKCS8PrivateKey(edPriv)
	if err != nil {
		t.Fatalf("序列化 ed25519 私钥失败: %v", err)
	}
	edKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: edDER})

	tests := []struct {
		name        string
		certPEM     []byte
		keyPEM      []byte
		wantErrText string // 为空表示应校验通过
	}{
		{name: "RSA 证书与 PKCS8 私钥匹配", certPEM: rsaCert, keyPEM: rsaPKCS8PEM},
		{name: "RSA 证书与 PKCS1 私钥匹配", certPEM: rsaCert, keyPEM: rsaPKCS1PEM},
		{name: "ECDSA 证书与 PKCS8 私钥匹配", certPEM: ecCert, keyPEM: ecPKCS8PEM},
		{name: "ECDSA 证书与 SEC1 私钥匹配", certPEM: ecCert, keyPEM: ecSEC1PEM},
		{name: "证书为空", certPEM: nil, keyPEM: rsaPKCS8PEM, wantErrText: "证书数据为空"},
		{name: "私钥为空", certPEM: rsaCert, keyPEM: nil, wantErrText: "私钥数据为空"},
		{name: "证书 PEM 非证书块", certPEM: rsaPKCS8PEM, keyPEM: rsaPKCS8PEM, wantErrText: "无效的证书格式"},
		{name: "证书内容不合法", certPEM: invalidCertContent, keyPEM: rsaPKCS8PEM, wantErrText: "无效的证书内容"},
		{name: "私钥 PEM 无块", certPEM: rsaCert, keyPEM: []byte("-----BEGIN BROKEN-----"), wantErrText: "无效的私钥格式"},
		{name: "私钥块类型不支持", certPEM: rsaCert, keyPEM: pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: rsaPKCS8PEM}), wantErrText: "不支持的私钥类型"},
		{name: "私钥内容不合法", certPEM: rsaCert, keyPEM: invalidKeyContent, wantErrText: "无效的私钥内容"},
		{name: "私钥算法不支持", certPEM: rsaCert, keyPEM: edKeyPEM, wantErrText: "不支持的私钥算法"},
		{name: "RSA 证书配另一把 RSA 私钥", certPEM: otherRSACert, keyPEM: rsaPKCS8PEM, wantErrText: "证书与私钥不匹配"},
		{name: "ECDSA 证书配另一把 ECDSA 私钥", certPEM: otherECCert, keyPEM: ecPKCS8PEM, wantErrText: "证书与私钥不匹配"},
		{name: "ECDSA 证书配 RSA 私钥", certPEM: ecCert, keyPEM: rsaPKCS8PEM, wantErrText: "证书与私钥不匹配"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCertAndKey(tt.certPEM, tt.keyPEM)
			if tt.wantErrText == "" {
				if err != nil {
					t.Fatalf("期望校验通过，实际返回错误: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("期望错误包含 %q，实际返回 nil", tt.wantErrText)
			}
			if !strings.Contains(err.Error(), tt.wantErrText) {
				t.Fatalf("错误信息应包含 %q，实际为 %q", tt.wantErrText, err.Error())
			}
		})
	}
}

func TestPublicKeysEqual(t *testing.T) {
	rsaA := newRSAKey(t, "PRIVATE KEY")
	rsaB := newRSAKey(t, "PRIVATE KEY")
	ecA := newECKey(t, "PRIVATE KEY")
	ecB := newECKey(t, "PRIVATE KEY")

	rsaPubA, ok := rsaA.signer.Public().(*rsa.PublicKey)
	if !ok {
		t.Fatal("rsaA 公钥类型断言失败")
	}
	rsaPubB, ok := rsaB.signer.Public().(*rsa.PublicKey)
	if !ok {
		t.Fatal("rsaB 公钥类型断言失败")
	}
	ecPubA, ok := ecA.signer.Public().(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("ecA 公钥类型断言失败")
	}
	ecPubB, ok := ecB.signer.Public().(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("ecB 公钥类型断言失败")
	}

	tests := []struct {
		name string
		a    crypto.PublicKey
		b    crypto.PublicKey
		want bool
	}{
		{name: "同类型同公钥", a: rsaPubA, b: rsaPubA, want: true},
		{name: "同为 RSA 但公钥不同", a: rsaPubA, b: rsaPubB, want: false},
		{name: "同为 ECDSA 同公钥", a: ecPubA, b: ecPubA, want: true},
		{name: "同为 ECDSA 但公钥不同", a: ecPubA, b: ecPubB, want: false},
		{name: "RSA 与 ECDSA 类型不一致", a: rsaPubA, b: ecPubA, want: false},
		{name: "左侧为 nil", a: nil, b: rsaPubA, want: false},
		{name: "右侧为 nil", a: rsaPubA, b: nil, want: false},
		{name: "不支持的公钥类型", a: "string-is-not-a-public-key", b: "string-is-not-a-public-key", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publicKeysEqual(tt.a, tt.b); got != tt.want {
				t.Fatalf("publicKeysEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCertificate(t *testing.T) {
	key := newRSAKey(t, "PRIVATE KEY")
	notBefore := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	notAfter := time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)
	certPEM := newSelfSignedCert(t, key, certTemplate("api.example.com", 12345, []string{"api.example.com"}, notBefore, notAfter))

	parsed, err := ParseCertificate(certPEM)
	if err != nil {
		t.Fatalf("解析证书失败: %v", err)
	}
	if parsed.Domain != "api.example.com" {
		t.Fatalf("Domain = %q, want %q", parsed.Domain, "api.example.com")
	}
	if parsed.SerialNumber != "12345" {
		t.Fatalf("SerialNumber = %q, want %q", parsed.SerialNumber, "12345")
	}
	if !parsed.NotBefore.Equal(notBefore) {
		t.Fatalf("NotBefore = %v, want %v", parsed.NotBefore, notBefore)
	}
	if !parsed.NotAfter.Equal(notAfter) {
		t.Fatalf("NotAfter = %v, want %v", parsed.NotAfter, notAfter)
	}
	if _, offset := parsed.NotAfter.Zone(); offset != 0 {
		t.Fatalf("NotAfter 应为 UTC，实际时区偏移 %d 秒", offset)
	}

	errCases := []struct {
		name        string
		certPEM     []byte
		wantErrText string
	}{
		{name: "证书为空", certPEM: nil, wantErrText: "证书数据为空"},
		{name: "PEM 非证书块", certPEM: key.keyPEM, wantErrText: "无效的证书格式"},
		{name: "证书内容不合法", certPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("bad")}), wantErrText: "解析证书失败"},
	}
	for _, ec := range errCases {
		t.Run(ec.name, func(t *testing.T) {
			if _, err := ParseCertificate(ec.certPEM); err == nil || !strings.Contains(err.Error(), ec.wantErrText) {
				t.Fatalf("期望错误包含 %q，实际为 %v", ec.wantErrText, err)
			}
		})
	}
}

func TestParseCertificateFull(t *testing.T) {
	now := time.Now()

	t.Run("RSA 通配符证书", func(t *testing.T) {
		key := newRSAKey(t, "PRIVATE KEY")
		certPEM := newSelfSignedCert(t, key, certTemplate(
			"example.com", 999, []string{"example.com", "*.example.com"},
			now.Add(-10*24*time.Hour), now.Add(80*24*time.Hour),
		))

		full, err := ParseCertificateFull(certPEM)
		if err != nil {
			t.Fatalf("解析证书失败: %v", err)
		}
		if full.Subject != "example.com" || full.Issuer != "example.com" {
			t.Fatalf("Subject/Issuer = %q/%q, want example.com/example.com", full.Subject, full.Issuer)
		}
		if full.SerialNumber != "999" {
			t.Fatalf("SerialNumber = %q, want 999", full.SerialNumber)
		}
		if len(full.DNSNames) != 2 {
			t.Fatalf("DNSNames = %v, want 2 项", full.DNSNames)
		}
		if !full.IsWildcard {
			t.Fatal("IsWildcard 应为 true")
		}
		if full.IsECC {
			t.Fatal("IsECC 应为 false")
		}
		if full.IsExpired {
			t.Fatal("IsExpired 应为 false")
		}
		if full.KeyLength != 2048 {
			t.Fatalf("KeyLength = %d, want 2048", full.KeyLength)
		}
		if full.Version != 3 {
			t.Fatalf("Version = %d, want 3", full.Version)
		}
		if full.PublicKeyAlgorithm != "RSA" {
			t.Fatalf("PublicKeyAlgorithm = %q, want RSA", full.PublicKeyAlgorithm)
		}
		if full.ExpiryDays < 79 || full.ExpiryDays > 80 {
			t.Fatalf("ExpiryDays = %d, 期望 79~80", full.ExpiryDays)
		}
		// 已过 10 天，总有效期 90 天 → 约 11%
		if full.ExpiryPercent < 5 || full.ExpiryPercent > 20 {
			t.Fatalf("ExpiryPercent = %d, 期望落在 5~20", full.ExpiryPercent)
		}
	})

	t.Run("ECDSA 证书", func(t *testing.T) {
		key := newECKey(t, "PRIVATE KEY")
		certPEM := newSelfSignedCert(t, key, certTemplate(
			"ec.example.com", 1000, []string{"ec.example.com"},
			now.Add(-time.Hour), now.Add(48*time.Hour),
		))

		full, err := ParseCertificateFull(certPEM)
		if err != nil {
			t.Fatalf("解析证书失败: %v", err)
		}
		if !full.IsECC {
			t.Fatal("IsECC 应为 true")
		}
		if full.KeyLength != 256 {
			t.Fatalf("KeyLength = %d, want 256", full.KeyLength)
		}
		if full.IsWildcard {
			t.Fatal("IsWildcard 应为 false")
		}
		if full.PublicKeyAlgorithm != "ECDSA" {
			t.Fatalf("PublicKeyAlgorithm = %q, want ECDSA", full.PublicKeyAlgorithm)
		}
	})

	t.Run("已过期证书：ExpiryPercent 被截断为 100 且 IsExpired 为 true", func(t *testing.T) {
		key := newRSAKey(t, "PRIVATE KEY")
		certPEM := newSelfSignedCert(t, key, certTemplate(
			"expired.example.com", 1001, nil,
			now.Add(-48*time.Hour), now.Add(-time.Hour),
		))

		full, err := ParseCertificateFull(certPEM)
		if err != nil {
			t.Fatalf("解析证书失败: %v", err)
		}
		if !full.IsExpired {
			t.Fatal("IsExpired 应为 true")
		}
		if full.ExpiryPercent != 100 {
			t.Fatalf("ExpiryPercent = %d, want 100", full.ExpiryPercent)
		}
		// 已过期 1 小时，ExpiryDays 为 int(负数小时/24) 截断为 0；只需保证不再为正数
		if full.ExpiryDays > 0 {
			t.Fatalf("ExpiryDays = %d, 期望不为正数", full.ExpiryDays)
		}
	})

	t.Run("尚未生效证书：ExpiryPercent 被截断为 0", func(t *testing.T) {
		key := newRSAKey(t, "PRIVATE KEY")
		certPEM := newSelfSignedCert(t, key, certTemplate(
			"future.example.com", 1002, nil,
			now.Add(24*time.Hour), now.Add(48*time.Hour),
		))

		full, err := ParseCertificateFull(certPEM)
		if err != nil {
			t.Fatalf("解析证书失败: %v", err)
		}
		if full.ExpiryPercent != 0 {
			t.Fatalf("ExpiryPercent = %d, want 0", full.ExpiryPercent)
		}
		if full.IsExpired {
			t.Fatal("IsExpired 应为 false")
		}
	})

	t.Run("非法输入", func(t *testing.T) {
		key := newRSAKey(t, "PRIVATE KEY")
		cases := []struct {
			name        string
			certPEM     []byte
			wantErrText string
		}{
			{name: "证书为空", certPEM: nil, wantErrText: "证书数据为空"},
			{name: "PEM 非证书块", certPEM: key.keyPEM, wantErrText: "无效的证书格式"},
			{name: "证书内容不合法", certPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("bad")}), wantErrText: "解析证书失败"},
		}
		for _, c := range cases {
			if _, err := ParseCertificateFull(c.certPEM); err == nil || !strings.Contains(err.Error(), c.wantErrText) {
				t.Fatalf("%s: 期望错误包含 %q，实际为 %v", c.name, c.wantErrText, err)
			}
		}
	})
}
