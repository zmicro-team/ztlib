package authorize

import (
	"context"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
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
	token, err := userAuthorize.GenerateToken(context.Background(),user)
	if err != nil {
		t.Errorf("Error generating token: %v", err)
	}
	if token == "" {
		t.Errorf("Expected non-empty token, got empty")
	}
}
func TestVerifyToken(t *testing.T) {
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
	tok, getUser, err := userAuthorize.VerifyToken(context.Background(), token)
	if err != nil {
		t.Errorf("Error verifying token: %v", err)
	}
	if tok == nil {
		t.Errorf("Expected non-nil token, got nil")
	}
	if getUser == nil {
		t.Errorf("Expected non-nil user, got nil")
	}
}
