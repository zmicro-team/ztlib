package rsax

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
)

func GetPKPubKey(pub []byte) (pub_rsa *rsa.PublicKey, err error) {
	return getPKPubKey(pub)
}

func getPKPubKey(pub []byte) (pubRsa *rsa.PublicKey, err error) {
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

// 公钥加密或解密byte
func pubKeyByte(pub *rsa.PublicKey, in []byte, isEncrypt bool) ([]byte, error) {
	k := (pub.N.BitLen() + 7) / 8
	if isEncrypt {
		k = k - 11
	}
	if len(in) <= k {
		if isEncrypt {
			return rsa.EncryptPKCS1v15(rand.Reader, pub, in)
		} else {
			return pubKeyDecrypt(pub, in)
		}
	} else {
		iv := make([]byte, k)
		out := bytes.NewBuffer(iv)
		if err := pubKeyIO(pub, bytes.NewReader(in), out, isEncrypt); err != nil {
			return nil, err
		}
		return io.ReadAll(out)
	}
}

// 公钥加密或解密Reader
func pubKeyIO(pub *rsa.PublicKey, in io.Reader, out io.Writer, isEncrypt bool) (err error) {
	k := (pub.N.BitLen() + 7) / 8
	if isEncrypt {
		k = k - 11
	}
	buf := make([]byte, k)
	var b []byte
	size := 0
	for {
		size, err = in.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if size < k {
			b = buf[:size]
		} else {
			b = buf
		}
		if isEncrypt {
			b, err = rsa.EncryptPKCS1v15(rand.Reader, pub, b)
		} else {
			b, err = pubKeyDecrypt(pub, b)
		}
		if err != nil {
			return err
		}
		if _, err = out.Write(b); err != nil {
			return err
		}
	}
}
