package aesx

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"math/rand"
	"time"
)

// AES-GCM 加密数据
func GCMEncrypt(originText, additional, key []byte) (nonce []byte, cipherText []byte, err error) {
	return gcmEncrypt(originText, additional, key)
}

// AES-GCM 解密数据
func GCMDecrypt(cipherText, nonce, additional, key []byte) ([]byte, error) {
	return gcmDecrypt(cipherText, nonce, additional, key)
}

func gcmDecrypt(secretData, nonce, additional, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM(),error:%w", err)
	}
	originByte, err := gcm.Open(nil, nonce, secretData, additional)
	if err != nil {
		return nil, err
	}
	return originByte, nil
}

// gcmEncrypt 使用AES-GCM算法对原始文本进行加密。
// originText: 要加密的原始文本。
// additional: 附加数据，用于提供额外的安全性。
// key: 加密密钥。
// 返回加密后的密文、随机生成的nonce和可能的错误。
func gcmEncrypt(originText, additional, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	nonce := []byte(RandomString(12))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("cipher.NewGCM(),error:%w", err)
	}
	cipherBytes := gcm.Seal(nil, nonce, originText, additional)
	return nonce, cipherBytes, nil
}

func RandomString(l int) string {
	const str = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, l)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range result {
		result[i] = str[r.Intn(len(str))]
	}
	return string(result)
}

// * 随机生成补全码
// func RandomString(l int) string {
// 	str := "0123456789AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"
// 	bytes := []byte(str)
// 	var result []byte = make([]byte, 0, l)
// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	for i := 0; i < l; i++ {
// 		result = append(result, bytes[r.Intn(len(bytes))])
// 	}
// 	return BytesToString(result)
// }

// & BytesToString 0 拷贝转换 slice byte 为 string
// func BytesToString(b []byte) (s string) {
// 	_bptr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
// 	_sptr := (*reflect.StringHeader)(unsafe.Pointer(&s))
// 	_sptr.Data = _bptr.Data
// 	_sptr.Len = _bptr.Len
// 	return s
// }
