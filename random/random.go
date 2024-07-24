package stringx

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultLetterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits     = 6 // 6 bits to represent a letter index
	idLen             = 8
	defaultRandLen    = 8
	letterIdxMask     = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax      = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = newLockedSource(time.Now().UnixNano())

type lockedSource struct {
	source rand.Source
	lock   sync.Mutex
}

func newLockedSource(seed int64) *lockedSource {
	return &lockedSource{
		source: rand.NewSource(seed),
	}
}

func (ls *lockedSource) Int63() int64 {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	return ls.source.Int63()
}

func (ls *lockedSource) Seed(seed int64) {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	ls.source.Seed(seed)
}

// Rand returns a random string.
func Rand() string {
	return Randn(defaultRandLen,"")
}

// RandId returns a random id string.
func RandId() string {
	b := make([]byte, idLen)
	_, err := crand.Read(b)
	if err != nil {
		return Randn(idLen, "")
	}

	return fmt.Sprintf("%x%x%x%x", b[0:2], b[2:4], b[4:6], b[6:8])
}

// Randn returns a random string with length n.
// Randn 生成指定长度的随机字符串。
// 参数 n 表示生成的字符串长度。
// 参数 letterBytes 是用于生成随机字符串的字符集。
// 如果 letterBytes 为空，则使用默认的字符集 defaultLetterBytes。
// 返回生成的随机字符串。
func Randn(n int, letterBytes string) string {
	// 如果 letterBytes 为空，则使用默认的字符集 defaultLetterBytes。
	if letterBytes == "" {
		letterBytes = defaultLetterBytes
	}

	b := make([]byte, n)
	// 使用 src.Int63() 生成 63 位的随机数，足够生成 letterIdxMax 个字符！
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		// 如果 remain 为 0，则重新生成随机数 cache，并重置 remain 为 letterIdxMax。
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		// 通过位运算获取 cache 的低 letterIdxBits 位，并将其作为索引获取 letterBytes 中的字符。
		// 如果索引小于 letterBytes 的长度，则将字符存入 b 中，并将索引递减。
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// Seed sets the seed to seed.
func Seed(seed int64) {
	src.Seed(seed)
}
