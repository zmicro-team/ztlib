package rsax

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

type (
	options struct {
		pubStr string // * 公钥字符串
		priStr string // * 私钥字符串
	}

	RSASecurityOption func(*options)

	RSASecurity struct {
		options options
		pubKey  *rsa.PublicKey  // * 公钥
		priKey  *rsa.PrivateKey // * 私钥
	}
)

func SetPublicString(val string) RSASecurityOption {
	return func(o *options) {
		o.pubStr = val
	}
}

func SetPrivateString(val string) RSASecurityOption {
	return func(o *options) {
		o.priStr = val
	}
}

func New(opts ...RSASecurityOption) *RSASecurity {
	options := &options{}
	for _, v := range opts {
		v(options)
	}
	ins := &RSASecurity{options: *options}
	if ins.options.pubStr != "" {
		pubKey, err := getPKPubKey([]byte(ins.options.pubStr))
		if err != nil {
			panic(err)
		}
		ins.pubKey = pubKey
	}

	if ins.options.priStr != "" {
		priKey, err := getPKPrivKey([]byte(ins.options.priStr))
		if err != nil {
			panic(err)
		}
		ins.priKey = priKey
	}

	return ins
}

// * 使用公钥加密
func (rs *RSASecurity) pubKeyEncode(input []byte) ([]byte, error) {
	if rs.pubKey == nil {
		return []byte(""), ErrNoPrivatekeySet
	}
	output := bytes.NewBuffer(nil)
	err := pubKeyIO(rs.pubKey, bytes.NewReader(input), output, true)
	if err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// * 公钥解密
func (rs *RSASecurity) pubKeyDecode(input []byte) ([]byte, error) {
	if rs.pubKey == nil {
		return []byte(""), ErrNoPrivatekeySet
	}
	output := bytes.NewBuffer(nil)
	err := pubKeyIO(rs.pubKey, bytes.NewReader(input), output, false)
	if err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// * 私钥加密
func (rs *RSASecurity) priKeyEncode(input []byte) ([]byte, error) {
	if rs.priKey == nil {
		return []byte(""), ErrNoPrivatekeySet
	}
	output := bytes.NewBuffer(nil)
	err := priKeyIO(rs.priKey, bytes.NewReader(input), output, true)
	if err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// * 私钥解密
func (rs *RSASecurity) priKeyDecode(input []byte) ([]byte, error) {
	if rs.priKey == nil {
		return []byte(""), ErrNoPrivatekeySet
	}
	output := bytes.NewBuffer(nil)
	err := priKeyIO(rs.priKey, bytes.NewReader(input), output, false)
	if err != nil {
		return []byte(""), err
	}

	return io.ReadAll(output)
}

// * 使用RSAWithMD5算法签名
func (rs *RSASecurity) SignMd5WithRsa(data string) (string, error) {
	md5Hash := md5.New()
	s_data := []byte(data)
	md5Hash.Write(s_data)
	hashed := md5Hash.Sum(nil)

	signByte, err := rsa.SignPKCS1v15(rand.Reader, rs.priKey, crypto.MD5, hashed)
	sign := base64.StdEncoding.EncodeToString(signByte)
	return string(sign), err
}

// * 使用RSAWithSHA1算法签名
func (rs *RSASecurity) SignSha1WithRsa(data string) (string, error) {
	sha1Hash := sha1.New()
	s_data := []byte(data)
	sha1Hash.Write(s_data)
	hashed := sha1Hash.Sum(nil)

	signByte, err := rsa.SignPKCS1v15(rand.Reader, rs.priKey, crypto.SHA1, hashed)
	sign := base64.StdEncoding.EncodeToString(signByte)
	return string(sign), err
}

// * 使用RSAWithSHA256算法签名
func (rs *RSASecurity) SignSha256WithRsa(data string) (string, error) {
	sha256Hash := sha256.New()
	s_data := []byte(data)
	sha256Hash.Write(s_data)
	hashed := sha256Hash.Sum(nil)

	signByte, err := rsa.SignPKCS1v15(rand.Reader, rs.priKey, crypto.SHA256, hashed)
	sign := base64.StdEncoding.EncodeToString(signByte)
	return string(sign), err
}

// * 使用RSAWithMD5验证签名
func (rs *RSASecurity) VerifySignMd5WithRsa(data string, signData string) error {
	sign, err := base64.StdEncoding.DecodeString(signData)
	if err != nil {
		return err
	}
	hash := md5.New()
	hash.Write([]byte(data))
	return rsa.VerifyPKCS1v15(rs.pubKey, crypto.MD5, hash.Sum(nil), sign)
}

// * 使用RSAWithSHA1验证签名
func (rs *RSASecurity) VerifySignSha1WithRsa(data string, signData string) error {
	sign, err := base64.StdEncoding.DecodeString(signData)
	if err != nil {
		return err
	}
	hash := sha1.New()
	hash.Write([]byte(data))
	return rsa.VerifyPKCS1v15(rs.pubKey, crypto.SHA1, hash.Sum(nil), sign)
}

// * 使用RSAWithSHA256验证签名
func (rs *RSASecurity) VerifySignSha256WithRsa(data string, signData string) error {
	sign, err := base64.StdEncoding.DecodeString(signData)
	if err != nil {
		return err
	}
	hash := sha256.New()
	hash.Write([]byte(data))
	return rsa.VerifyPKCS1v15(rs.pubKey, crypto.SHA256, hash.Sum(nil), sign)
}
