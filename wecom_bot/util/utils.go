package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

func JsonMarshal(data any) string {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// DecodeBase64ToReader 将 Base64 编码的字符串解码，并返回一个包含解码数据的 io.Reader。
// 这个 io.Reader 允许你像读取文件一样读取内存中的解码数据，但不会实际创建文件。
//
// 参数:
//
//	base64String: Base64 编码的输入字符串。
//
// 返回值:
//
//	io.Reader: 一个可以读取解码后数据的读取器。
//	error:     如果解码失败，则返回一个 error。成功则返回 nil。
func DecodeBase64ToReader(base64String string) (io.Reader, error) {
	// 1. 解码 Base64 字符串
	decodedBytes, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, fmt.Errorf("解码 Base64 字符串失败: %w", err) // 返回 nil reader 和错误
	}

	// 2. 使用解码后的字节创建一个 bytes.Reader
	// bytes.NewReader 返回一个 *bytes.Reader，它实现了 io.Reader, io.Seeker, io.ReaderAt 等接口。
	reader := bytes.NewReader(decodedBytes)

	// 3. 返回创建的 reader 和 nil 错误
	return reader, nil
}
