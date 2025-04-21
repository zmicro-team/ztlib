package authorize

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
)

type AuthorizeConfig struct {
	Expire time.Duration
	Issuer string

	SignatureAlgorithm jwa.SignatureAlgorithm // 签名算法
	SecretKey          string                 // 签名密钥

	KeySignatureAlgorithm jwa.KeyEncryptionAlgorithm // 键密钥签名算法
	PublicKeyPath         string
	PrivateKeyPath        string

	RefreshTimeout time.Duration
}
