package authorize

import (
	"context"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// IAuthorize 授权接口
type IAuthorize interface {
	GenerateToken(ctx context.Context, user IAuthorizeOther) (str string, err error)
	VerifyToken(ctx context.Context, token string, user IAuthorizeOther) (jwt.Token, error)
	// RefreshToken(ctx context.Context, token string) (str string, err error)
}

// IAuthorizeOther token中加密的数据
type IAuthorizeOther interface {
	Encrypt(ctx context.Context, algorithm jwa.KeyEncryptionAlgorithm, jwkRSAPublicKey jwk.Key) (string, error)

	Decrypt(ctx context.Context, encrypted string, algorithm jwa.KeyEncryptionAlgorithm, jwkRSAPrivateKey jwk.Key) error

	GetId(context.Context) string

	GetBan(context.Context) bool

	SetBan(context.Context, bool)
}
