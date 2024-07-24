package pk

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestGetPKPrivateKey(t *testing.T) {
	// 生成一个 RSA 私钥
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	// 将私钥编码为 PKCS1 格式
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	// 创建一个 PEM 结构体
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})
	// 使用 GetPKPrivateKey 函数解析私钥
	_, err := GetPKPrivateKey(privPem)
	if err != nil {
		t.Errorf("Failed to parse private key: %v", err)
	}
}

func TestGetPKPublicKey(t *testing.T) {
	// 生成一个 RSA 私钥
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	// 从私钥中获取公钥
	pubKey := &privKey.PublicKey
	// 将公钥编码为 PKIX 格式
	pubKeyBytes, _ := x509.MarshalPKIXPublicKey(pubKey)
	// 创建一个 PEM 结构体
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	// 使用 GetPKPubKey 函数解析公钥
	_, err := GetPKPublicKey(pubPem)
	if err != nil {
		t.Errorf("Failed to parse public key: %v", err)
	}
}
