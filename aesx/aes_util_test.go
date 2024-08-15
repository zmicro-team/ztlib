package aesx

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLENCODE(t *testing.T) {
	unix := "123"
	_ = unix
	appid := "100000"
	_ = appid
	// sign  := "123"
	value := url.Values{}
	value.Add("grant_type", "123")
	value.Add("code", "123")
	value.Add("refresh_token", "123")
	value.Add("b", "123")
	value.Add("a", "123")
	code := value.Encode()
	_ = code
	t.Log(code)
	// cIjeYALQGUFrR1fFVWA0+g==
	sign, err := ECBEncryptByUrlBase64([]byte(fmt.Sprintf("%s", code)), []byte("cIjeYALQGUFrR1fFVWA0+g=="))
	assert.NoError(t, err)
	t.Log(string(sign))
	data, err := ECBDecryptByUrlBase64(sign, []byte("cIjeYALQGUFrR1fFVWA0+g=="))
	assert.NoError(t, err)
	_ = data
	t.Log(string(data))
}
