package rsax

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

func GetPKPrivKey(priv []byte) (priv_key *rsa.PrivateKey, err error) {
	return getPKPrivKey(priv)
}

func getPKPrivKey(priv []byte) (privKey *rsa.PrivateKey, err error) {
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

// 私钥加密或解密byte
func priKeyByte(pri *rsa.PrivateKey, in []byte, isEncrypt bool) ([]byte, error) {
	k := (pri.N.BitLen() + 7) / 8
	if isEncrypt {
		k = k - 11
	}
	if len(in) <= k {
		if isEncrypt {
			return priKeyEncrypt(rand.Reader, pri, in)
		} else {
			return rsa.DecryptPKCS1v15(rand.Reader, pri, in)
		}
	} else {
		iv := make([]byte, k)
		out := bytes.NewBuffer(iv)
		if err := priKeyIO(pri, bytes.NewReader(in), out, isEncrypt); err != nil {
			return nil, err
		}
		return io.ReadAll(out)
	}
}

// * 私钥加密或解密Reader
func priKeyIO(pri *rsa.PrivateKey, r io.Reader, w io.Writer, isEncrypt bool) (err error) {
	k := (pri.N.BitLen() + 7) / 8
	if isEncrypt {
		k = k - 11
	}
	buf := make([]byte, k)
	var b []byte
	size := 0
	for {
		size, err = r.Read(buf)
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
			b, err = priKeyEncrypt(rand.Reader, pri, b)
		} else {
			b, err = rsa.DecryptPKCS1v15(rand.Reader, pri, b)
		}
		if err != nil {
			return err
		}
		if _, err = w.Write(b); err != nil {
			return err
		}
	}
}
