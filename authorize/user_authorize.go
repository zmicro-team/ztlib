package authorize

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/zmicro-team/ztlib/authorize/pk"
)

const UserAuthorizeInfo = "userAuthorizeInfo"

type UserAuthorize struct {
	publicKey  jwk.Key
	privateKey jwk.Key
	secretKey  jwk.Key
	options    *AuthorizeConfig
	banAccount func(context.Context, IAuthorizeOther) bool // 是否已经停用
}

func NewUserAuthorize(options *AuthorizeConfig) *UserAuthorize {
	if options.KeySignatureAlgorithm == "" {
		options.KeySignatureAlgorithm = jwa.RSA_OAEP
	}
	if options.SignatureAlgorithm == "" {
		options.SignatureAlgorithm = jwa.HS256
	}
	if options.Expire == 0 {
		options.Expire = time.Hour * 24
	}
	if options.RefreshTimeout == 0 {
		options.RefreshTimeout = time.Hour * 24
	}
	userAuthorize := &UserAuthorize{options: options}

	privateKeyBytes, err := os.ReadFile(options.PrivateKeyPath)
	if err != nil {
		panic(err)
	}
	privateKey, err := pk.GetPKPrivateKey(privateKeyBytes)
	if err != nil {
		panic(err)
	}
	if key, err := jwk.FromRaw(privateKey); err != nil {
		panic(err)
	} else {
		userAuthorize.privateKey = key
	}
	publicKeyBytes, err := os.ReadFile(options.PublicKeyPath)
	if err != nil {
		panic(err)
	}
	publicKey, err := pk.GetPKPublicKey(publicKeyBytes)
	if err != nil {
		panic(err)
	}
	if key, err := jwk.FromRaw(publicKey); err != nil {
		panic(err)
	} else {
		userAuthorize.publicKey = key
	}
	if key, err := jwk.FromRaw([]byte(options.SecretKey)); err != nil {
		panic(err)
	} else {
		userAuthorize.secretKey = key
	}
	return userAuthorize
}

// SetBanAccount
func (userAuthorize *UserAuthorize) SetBanAccount(callback func(context.Context, IAuthorizeOther) bool) {
	userAuthorize.banAccount = callback
}

func (userAuthorize *UserAuthorize) GenerateToken(ctx context.Context, user IAuthorizeOther) (str string, err error) {
	userEncrypt, err := user.Encrypt(ctx, userAuthorize.options.KeySignatureAlgorithm, userAuthorize.privateKey)
	if err != nil {
		return "", err
	}
	jwtToken, err := jwt.NewBuilder().
		Issuer(userAuthorize.options.Issuer).
		Expiration(time.Now().Add(userAuthorize.options.Expire)).
		IssuedAt(time.Now()).
		NotBefore(time.Now()).
		Claim(UserAuthorizeInfo, userEncrypt).
		Build()
	if err != nil {
		return "", err
	}
	signed, err := jwt.Sign(jwtToken, jwt.WithKey(userAuthorize.options.SignatureAlgorithm, userAuthorize.secretKey))
	if err != nil {
		return "", err
	}
	return string(signed), nil
}

func (userAuthorize *UserAuthorize) VerifyToken(ctx context.Context, token string, user IAuthorizeOther) (jwt.Token, error) {
	jwtToken, err := jwt.ParseString(token, jwt.WithKey(userAuthorize.options.SignatureAlgorithm, userAuthorize.secretKey))
	if err != nil {
		return nil, err
	}
	claims := jwtToken.PrivateClaims()
	userEncrypt, ok := claims[UserAuthorizeInfo]
	if !ok {
		return nil, err
	}
	err = user.Decrypt(ctx, userEncrypt.(string), userAuthorize.options.KeySignatureAlgorithm, userAuthorize.privateKey)
	if err != nil {
		return nil, err
	}
	if userAuthorize.banAccount != nil {
		user.SetBan(ctx, userAuthorize.banAccount(ctx, user))
	}
	return jwtToken, nil
}

type UserAuthorizeOther struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Ban  bool   `json:"ban"`
}

var _defUserAuthorizeOther = &UserAuthorizeOther{}

func (ua *UserAuthorizeOther) Encrypt(ctx context.Context, algorithm jwa.KeyEncryptionAlgorithm, jwkRSAPublicKey jwk.Key) (string, error) {
	payload, err := json.Marshal(ua)
	if err != nil {
		return "", err
	}
	encrypted, err := jwe.Encrypt([]byte(payload), jwe.WithKey(algorithm, jwkRSAPublicKey))
	if err != nil {
		return "", err
	}
	return string(encrypted), nil
}

func (ua *UserAuthorizeOther) Decrypt(ctx context.Context, encrypted string, algorithm jwa.KeyEncryptionAlgorithm, jwkRSAPrivateKey jwk.Key) error {
	decrypted, err := jwe.Decrypt([]byte(encrypted), jwe.WithKey(algorithm, jwkRSAPrivateKey))
	if err != nil {
		return err
	}
	err = json.Unmarshal(decrypted, &ua)
	if err != nil {
		return err
	}
	return nil
}

func (ua *UserAuthorizeOther) SetBan(ctx context.Context, b bool) {
	ua.Ban = b
}

func (ua *UserAuthorizeOther) GetBan(ctx context.Context) bool {
	return ua.Ban
}

func (ua *UserAuthorizeOther) GetId(ctx context.Context) string {
	return ua.Id
}

func (ua *UserAuthorizeOther) WithContextValue(ctx context.Context) context.Context {
	return context.WithValue(ctx, _defUserAuthorizeOther, ua)
}

func (ua *UserAuthorizeOther) GetFromContext(ctx context.Context) *UserAuthorizeOther {
	if ctx == nil {
		return nil
	}
	if rv := ctx.Value(_defUserAuthorizeOther); rv != nil {
		if v, ok := rv.(*UserAuthorizeOther); ok {
			return v
		}
	}
	return nil
}

func UserAuthorizeFromContext(ctx context.Context) *UserAuthorizeOther {
	if ctx == nil {
		return nil
	}
	if rv := ctx.Value(_defUserAuthorizeOther); rv != nil {
		if v, ok := rv.(*UserAuthorizeOther); ok {
			return v
		}
	}
	return nil
}

func GetUserIdFromContext(ctx context.Context) int64 {
	if ctx == nil {
		panic("ctx is nil")
	}
	if rv := ctx.Value(_defUserAuthorizeOther); rv != nil {
		if v, ok := rv.(*UserAuthorizeOther); ok {
			id, err := strconv.ParseInt(v.Id, 10, 64)
			if err != nil {
				panic(err)
			}
			return id
		}
	}
	panic("ctx value is nil")
}
