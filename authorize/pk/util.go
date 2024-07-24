package pk

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// GetPKPrivateKey 是一个函数，它接受一个字节切片作为私钥，并返回一个 rsa.PrivateKey 指针和一个错误。
func GetPKPrivateKey(priv []byte) (privKey *rsa.PrivateKey, err error) {
	// pem.Decode 将输入的字节切片解码为一个 PEM 结构体和剩余的输入。
	block, _ := pem.Decode(priv)
	// 如果解码后的 PEM 结构体为空，则返回一个错误。
	if block == nil {
		return nil, errors.New("priv key error")
	}
	// x509.ParsePKCS1PrivateKey 尝试解析输入的字节切片为 PKCS#1 私钥。
	privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	// 如果没有错误，即成功解析为 PKCS#1 私钥，则返回私钥和 nil 错误。
	if err == nil {
		return
	}
	// 如果 PKCS#1 解析失败，尝试解析输入的字节切片为 PKCS#8 私钥。
	privAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	// 如果 PKCS#8 解析失败，则返回 nil 私钥和错误。
	if err != nil {
		return
	}
	// 如果 PKCS#8 解析成功，将解析出的私钥转换为 rsa.PrivateKey 类型，并返回。
	privKey = privAny.(*rsa.PrivateKey)
	return
}

// GetPKPubKey 是一个函数，它接受一个字节切片作为公钥，并返回一个 rsa.PublicKey 指针和一个错误。
func GetPKPublicKey(pub []byte) (pubRsa *rsa.PublicKey, err error) {
	// pem.Decode 将输入的字节切片解码为一个 PEM 结构体和剩余的输入。
	block, _ := pem.Decode(pub)
	// x509.ParsePKCS1PublicKey 尝试解析输入的字节切片为 PKCS#1 公钥。
	pubRsa, err = x509.ParsePKCS1PublicKey(block.Bytes)
	// 如果没有错误，即成功解析为 PKCS#1 公钥，则返回公钥和 nil 错误。
	if err == nil {
		return
	}
	// 如果 PKCS#1 解析失败，尝试解析输入的字节切片为 PKIX 公钥。
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	// 如果 PKIX 解析失败，则返回 nil 公钥和错误。
	if err != nil {
		return
	}
	// 如果 PKIX 解析成功，将解析出的公钥转换为 rsa.PublicKey 类型，并返回。
	pubRsa = pubAny.(*rsa.PublicKey)
	return
}
