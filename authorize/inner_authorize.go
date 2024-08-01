package authorize

import (
	"context"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const InnerAuthorizeInfo = "innerAuthorizeInfo"

type InnerAuthorizeConfig struct {
	Secret string // 密钥
}

// InnerAuthorize 用于内部授权
type InnerAuthorize struct {
	secret                string                     // 密钥
	signatureAlgorithm    jwa.SignatureAlgorithm     // token 签名算法
	keySignatureAlgorithm jwa.KeyEncryptionAlgorithm // key 加密算法
	key                   jwk.Key
}

func NewInnerAuthorize(option *InnerAuthorizeConfig) *InnerAuthorize {
	key, err := jwk.FromRaw([]byte(option.Secret))
	if err != nil {
		panic(err)
	}
	auth := &InnerAuthorize{
		key:                   key,
		keySignatureAlgorithm: jwa.A128KW,
		signatureAlgorithm:    jwa.HS256,
		secret:                option.Secret}
	return auth
}

func (innerAuthorize *InnerAuthorize) GenerateToken(ctx context.Context, user IAuthorizeOther) (str string, err error) {
	userEncrypt, err := user.Encrypt(ctx, innerAuthorize.keySignatureAlgorithm, innerAuthorize.key)
	if err != nil {
		return "", err
	}
	jwtToken, err := jwt.NewBuilder().
		Issuer(InnerAuthorizeInfo).
		Claim(InnerAuthorizeInfo, userEncrypt).
		Build()
	if err != nil {
		return "", err
	}
	signed, err := jwt.Sign(jwtToken, jwt.WithKey(innerAuthorize.signatureAlgorithm, innerAuthorize.key))
	if err != nil {
		return "", err
	}
	return string(signed), nil
}

func (innerAuthorize *InnerAuthorize) VerifyToken(ctx context.Context, token string, user IAuthorizeOther) (jwt.Token, error) {
	jwtToken, err := jwt.ParseString(token, jwt.WithKey(innerAuthorize.signatureAlgorithm, innerAuthorize.key))
	if err != nil {
		return nil, err
	}
	claims := jwtToken.PrivateClaims()
	userEncrypt, ok := claims[InnerAuthorizeInfo]
	if !ok {
		return nil, err
	}
	err = user.Decrypt(ctx, userEncrypt.(string), innerAuthorize.keySignatureAlgorithm, innerAuthorize.key)
	if err != nil {
		return nil, err
	}
	return jwtToken, nil
}
