package authorize

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
)

type AuthorizeConfig struct {
	Expire                time.Duration
	Issuer                string
	KeySignatureAlgorithm jwa.KeyEncryptionAlgorithm
	PrivateKeyPath        string
	PublicKeyPath         string
	RefreshTimeout        time.Duration
	SecretKey             string
	SignatureAlgorithm    jwa.SignatureAlgorithm
}
