package authorize

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
)

type AuthorizeConfig struct {
	Expire                time.Duration
	Issuer                string
	SecretKey             string // 密钥
	PublicKeyPath         string
	RefreshTimeout        time.Duration
	PrivateKeyPath        string
	SignatureAlgorithm    jwa.SignatureAlgorithm     // 签名算法
	KeySignatureAlgorithm jwa.KeyEncryptionAlgorithm // 密钥签名算法
}
