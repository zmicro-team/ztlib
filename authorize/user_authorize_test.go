package authorize

import (
	"context"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
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
