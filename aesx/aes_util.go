package aesx

import (
	"encoding/base64"
	"net/url"
)

// map转换
func MapToURLValues(originData map[string]string) []byte {
	uv := url.Values{}
	for k, v := range originData {
		uv.Add(k, v)
	}
	return []byte(uv.Encode())
}

// AES-ECB 加密数据
func ECBEncryptByUrlBase64(originData, key []byte) ([]byte, error) {
	code, err := ecbEncrypt(originData, key)
	if err != nil {
		return nil, err
	}
	return []byte(base64.RawURLEncoding.EncodeToString(code)), nil
}

// AES-ECB 解密数据
func ECBDecryptByUrlBase64(secretData, key []byte) ([]byte, error) {
	secretData, err := base64.RawURLEncoding.DecodeString(string(secretData))
	if err != nil {
		return nil, err
	}
	return ecbDecrypt(secretData, key)
}
