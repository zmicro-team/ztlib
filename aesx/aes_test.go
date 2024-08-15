package aesx

import (
	"encoding/base64"
	"fmt"
	"testing"
)

var (
	secretKey = "JYRn4wbCy8KgVIZJaPhYTcTn2zixVC4Y"
	iv        = "JR3unO2glQuMhUx3"
)

func TestAesECBEncryptDecrypt(t *testing.T) {
	originData := "www.uc1024.cn"
	fmt.Println("originData:", originData)
	encryptData, err := ECBEncrypt([]byte(originData), []byte(secretKey))
	if err != nil {
		fmt.Println("AesCBCEncryptToString:", err)
		return
	}
	fmt.Println("encryptData:", string(encryptData))
	origin, err := ECBDecrypt(encryptData, []byte(secretKey))
	if err != nil {
		fmt.Println("AesDecryptToBytes:", err)
		return
	}
	fmt.Println("origin:", string(origin))
}

func TestAesCBCEncryptDecrypt(t *testing.T) {
	originData := "www.acfun.cn"
	fmt.Println("originData:", originData)
	encryptData, err := CBCEncrypt([]byte(originData), []byte(secretKey), []byte(iv))
	if err != nil {
		fmt.Println("CBCEncrypt:", err)
		return
	}
	fmt.Println("encryptData:", string(encryptData))
	origin, err := CBCDecrypt(encryptData, []byte(secretKey), []byte(iv))
	if err != nil {
		fmt.Println("CBCDecrypt:", err)
		return
	}
	fmt.Println("origin:", string(origin))
}

func TestEncryptGCM(t *testing.T) {
	data := `我是要加密的数据`
	additional := "transaction"
	apiV3key := "Cj5xC9RXf0GFCKWeD9PyY1ZWLgionbvx"
	fmt.Println("原始数据：", data)
	// 加密
	nonce, ciphertext, err := GCMEncrypt([]byte(data), []byte(additional), []byte(apiV3key))
	if err != nil {
		fmt.Println(err)
		return
	}
	encryptText := base64.StdEncoding.EncodeToString(ciphertext)
	fmt.Println("加密后：", encryptText)
	fmt.Println("nonce:", string(nonce))

	// 解密
	cipherBytes, _ := base64.StdEncoding.DecodeString(encryptText)
	decryptBytes, err := GCMDecrypt(cipherBytes, nonce, []byte(additional), []byte(apiV3key))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("解密后：", string(decryptBytes))
}
