package authorize

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
)

var testDefaultConfig = AuthorizeConfig{
	Expire:                time.Hour * 24,
	Issuer:                "example.com",
	KeySignatureAlgorithm: jwa.RSA_OAEP,
	PrivateKeyPath:        "./tools/rsa-private.key",
	PublicKeyPath:         "./tools/rsa-public.key",
	RefreshTimeout:        time.Hour * 24,
	SecretKey:             "secret",
	SignatureAlgorithm:    jwa.HS256,
}

func TestNewUserAuthorize(t *testing.T) {
	options := &testDefaultConfig
	userAuthorize := NewUserAuthorize(options)
	if userAuthorize.options != options {
		t.Errorf("Expected options to be set correctly")
	}
}

func TestGenerateToken(t *testing.T) {
	options := &testDefaultConfig
	userAuthorize := NewUserAuthorize(options)
	user := &UserAuthorizeOther{
		Id:   "123",
		Type: "user",
		Name: "John Doe",
	}
	token, err := userAuthorize.GenerateToken(context.Background(), user)
	assert.NoError(t, err)
	assert.NotEqual(t, token, "")
}
func TestVerifyTokenExpire(t *testing.T) {
	options := &testDefaultConfig
	options.Expire = time.Second
	userAuthorize := NewUserAuthorize(options)
	user := &UserAuthorizeOther{
		Id:   "123",
		Type: "user",
		Name: "John Doe",
	}
	token, err := userAuthorize.GenerateToken(context.Background(), user)
	if err != nil {
		t.Errorf("Error generating token: %v", err)
	}
	if token == "" {
		t.Errorf("Expected non-empty token, got empty")
	}
	time.Sleep(time.Second * 2)
	getUser := new(UserAuthorizeOther)
	tok, err := userAuthorize.VerifyToken(context.Background(), token, getUser)
	assert.NotNil(t, err)
	assert.Nil(t, tok)
	assert.Empty(t, getUser.Id)
}

func TestVerifyTokenNotExpire(t *testing.T) {
	options := &testDefaultConfig
	options.Expire = time.Second * 10
	userAuthorize := NewUserAuthorize(options)
	user := &UserAuthorizeOther{
		Id:   "123",
		Type: "user",
		Name: "John Doe",
	}
	token, err := userAuthorize.GenerateToken(context.Background(), user)
	if err != nil {
		t.Errorf("Error generating token: %v", err)
	}
	getUser := new(UserAuthorizeOther)
	tok, err := userAuthorize.VerifyToken(context.Background(), token, getUser)
	assert.NoError(t, err)
	assert.NotNil(t, tok)
	assert.NotNil(t, getUser)
}

func TestUserAuthorize_SetBanAccount(t *testing.T) {
	type args struct {
		callback func(context.Context, IAuthorizeOther) bool
	}
	tests := []struct {
		name          string
		userAuthorize *UserAuthorize
		args          args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.userAuthorize.SetBanAccount(tt.args.callback)
		})
	}
}

func TestUserAuthorize_GenerateToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		user IAuthorizeOther
	}
	tests := []struct {
		name          string
		userAuthorize *UserAuthorize
		args          args
		wantStr       string
		wantErr       bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, err := tt.userAuthorize.GenerateToken(tt.args.ctx, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserAuthorize.GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStr != tt.wantStr {
				t.Errorf("UserAuthorize.GenerateToken() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestUserAuthorize_VerifyToken(t *testing.T) {
	type args struct {
		ctx   context.Context
		token string
		user  IAuthorizeOther
	}
	tests := []struct {
		name          string
		userAuthorize *UserAuthorize
		args          args
		want          jwt.Token
		wantErr       bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.userAuthorize.VerifyToken(tt.args.ctx, tt.args.token, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserAuthorize.VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserAuthorize.VerifyToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserAuthorizeOther_Encrypt(t *testing.T) {
	type args struct {
		ctx             context.Context
		algorithm       jwa.KeyEncryptionAlgorithm
		jwkRSAPublicKey jwk.Key
	}
	tests := []struct {
		name    string
		ua      *UserAuthorizeOther
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ua.Encrypt(tt.args.ctx, tt.args.algorithm, tt.args.jwkRSAPublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserAuthorizeOther.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UserAuthorizeOther.Encrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserAuthorizeOther_Decrypt(t *testing.T) {
	type args struct {
		ctx              context.Context
		encrypted        string
		algorithm        jwa.KeyEncryptionAlgorithm
		jwkRSAPrivateKey jwk.Key
	}
	tests := []struct {
		name    string
		ua      *UserAuthorizeOther
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ua.Decrypt(tt.args.ctx, tt.args.encrypted, tt.args.algorithm, tt.args.jwkRSAPrivateKey); (err != nil) != tt.wantErr {
				t.Errorf("UserAuthorizeOther.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserAuthorizeOther_SetBan(t *testing.T) {
	type args struct {
		ctx context.Context
		b   bool
	}
	tests := []struct {
		name string
		ua   *UserAuthorizeOther
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ua.SetBan(tt.args.ctx, tt.args.b)
		})
	}
}

func TestUserAuthorizeOther_GetBan(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		ua   *UserAuthorizeOther
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ua.GetBan(tt.args.ctx); got != tt.want {
				t.Errorf("UserAuthorizeOther.GetBan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserAuthorizeOther_GetId(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		ua   *UserAuthorizeOther
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ua.GetId(tt.args.ctx); got != tt.want {
				t.Errorf("UserAuthorizeOther.GetId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserAuthorizeOther_WithContextValue(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		ua   *UserAuthorizeOther
		args args
		want context.Context
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ua.WithContextValue(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserAuthorizeOther.WithContextValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserAuthorizeOther_GetFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		ua   *UserAuthorizeOther
		args args
		want *UserAuthorizeOther
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ua.GetFromContext(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserAuthorizeOther.GetFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserAuthorizeFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want *UserAuthorizeOther
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UserAuthorizeFromContext(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserAuthorizeFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserIdFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserIdFromContext(tt.args.ctx); got != tt.want {
				t.Errorf("GetUserIdFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
