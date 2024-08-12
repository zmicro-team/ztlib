package rsax

import (
	"encoding/base64"
	"encoding/hex"
	"sort"
)

// * 公钥加密
func PublicEncrypt(data, publicKey string) (string, error) {
	gRsa := New(SetPublicString(publicKey))
	rsaData, err := gRsa.pubKeyEncode([]byte(data))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(rsaData), nil
}

// * 私钥加密
func PriKeyEncrypt(data, privateKey string) (string, error) {
	gRsa := New(SetPrivateString(privateKey))
	rsaData, err := gRsa.priKeyEncode([]byte(data))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(rsaData), nil
}

// * 公钥解密
func PublicDecrypt(data, publicKey string) (string, error) {
	dataBs, _ := base64.StdEncoding.DecodeString(data)
	gRsa := New(SetPublicString(publicKey))
	rsaData, err := gRsa.pubKeyDecode(dataBs)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(rsaData), nil
}

// * 私钥解密
func PriKeyDecrypt(data, privateKey string) (string, error) {
	dataBs, _ := base64.StdEncoding.DecodeString(data)
	gRsa := New(SetPrivateString(privateKey))
	rsaData, err := gRsa.priKeyDecode(dataBs)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(rsaData), nil
}

// * 使用RSAWithMD5算法签名
func SignMd5WithRsa(data string, privateKey string) (string, error) {
	gRsa := New(SetPrivateString(privateKey))
	sign, err := gRsa.SignMd5WithRsa(data)
	if err != nil {
		return "", err
	}
	return sign, err
}

// * 使用RSAWithSHA1算法签名
func SignSha1WithRsa(data string, privateKey string) (string, error) {
	gRsa := New(SetPrivateString(privateKey))

	sign, err := gRsa.SignSha1WithRsa(data)
	if err != nil {
		return "", err
	}

	return sign, err
}

// * 使用RSAWithSHA256算法签名
func SignSha256WithRsa(data string, privateKey string) (string, error) {
	gRsa := New(SetPrivateString(privateKey))
	sign, err := gRsa.SignSha256WithRsa(data)
	if err != nil {
		return "", err
	}
	return sign, err
}

// * 使用RSAWithMD5验证签名
func VerifySignMd5WithRsa(data string, signData string, publicKey string) error {
	gRsa := New(SetPublicString(publicKey))
	return gRsa.VerifySignMd5WithRsa(data, signData)
}

// * 使用RSAWithSHA1验证签名
func VerifySignSha1WithRsa(data string, signData string, publicKey string) error {
	gRsa := New(SetPublicString(publicKey))
	return gRsa.VerifySignSha1WithRsa(data, signData)
}

// * 使用RSAWithSHA256验证签名
func VerifySignSha256WithRsa(data string, signData string, publicKey string) error {
	gRsa := New(SetPublicString(publicKey))
	return gRsa.VerifySignSha256WithRsa(data, signData)
}

// * 根据key的排序输出拼接value
func MapSortToVal[T any](m map[string]T) []T {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i][0] < keys[j][0]
	})
	value := make([]T, len(m))
	for i := 0; i < len(keys); i++ {
		value[i] = m[keys[i]]
	}
	return value
}
