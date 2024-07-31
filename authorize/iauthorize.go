package authorize

import (
	"context"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type IAuthorizeOther interface {
	Encrypt(ctx context.Context, algorithm jwa.KeyEncryptionAlgorithm, jwkRSAPublicKey jwk.Key) (string, error)

	Decrypt(ctx context.Context, encrypted string, algorithm jwa.KeyEncryptionAlgorithm, jwkRSAPrivateKey jwk.Key) error

	GetId(context.Context) string

	GetBan(context.Context) bool

	SetBan(context.Context, bool)
}
