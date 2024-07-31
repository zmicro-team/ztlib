package extractor

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractors(t *testing.T) {
	ctx := context.WithValue(context.Background(), "Authorization", "Bearer abc123")
	ctx = context.WithValue(ctx, "Token", "abc123")
	request := &http.Request{
		Header: http.Header{
			"Authorization": []string{"Bearer xyz456"},
			"Token":         []string{"abc123"},
		},
	}

	headerExtractor := HeaderExtractor{"Token", "Authorization"}
	bearerExtractor := BearerExtractor{}
	authorizationExtractor := AuthorizationExtractor{}
	tokenExtractor := TokenExtractor{}

	// Test HeaderExtractor
	token, err := headerExtractor.Extract(ctx)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)

	token, err = headerExtractor.ExtractRequest(request)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)

	// Test BearerExtractor
	token, err = bearerExtractor.Extract(ctx)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)

	token, err = bearerExtractor.ExtractRequest(request)
	assert.Equal(t, err, nil)
	assert.Equal(t, "xyz456", token)

	// Test AuthorizationExtractor
	token, err = authorizationExtractor.Extract(ctx)
	assert.Equal(t, err, nil)
	assert.Equal(t, "Bearer abc123", token)

	token, err = authorizationExtractor.ExtractRequest(request)
	assert.Equal(t, err, nil)
	assert.Equal(t, "Bearer xyz456", token)

	// Test TokenExtractor
	token, err = tokenExtractor.Extract(ctx)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)

	token, err = tokenExtractor.ExtractRequest(request)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)
}
func TestHeaderExtractor_Extract(t *testing.T) {
	ctx := context.WithValue(context.Background(), "Authorization", "abc123")
	ctx = context.WithValue(ctx, "token", "321abc")
	extractor := HeaderExtractor{"Authorization", "Token"}

	token, err := extractor.Extract(ctx)

	assert.Equal(t, nil, err)
	assert.Equal(t, "abc123", token)
}

func TestBearerExtractor_Extract(t *testing.T) {
	ctx := context.WithValue(context.Background(), "Authorization", "Bearer abc123")
	extractor := BearerExtractor{}

	token, err := extractor.Extract(ctx)

	assert.Equal(t, nil, err)
	assert.Equal(t, "abc123", token)
}

func TestAuthorizationExtractor_Extract(t *testing.T) {
	ctx := context.WithValue(context.Background(), "Authorization", "Bearer abc123")
	extractor := AuthorizationExtractor{}

	token, err := extractor.Extract(ctx)

	assert.Equal(t, nil, err)
	assert.Equal(t, "Bearer abc123", token)
}

func TestTokenExtractor_Extract(t *testing.T) {
	ctx := context.WithValue(context.Background(), "Token", "abc123")
	extractor := TokenExtractor{}

	token, err := extractor.Extract(ctx)

	assert.Equal(t, nil, err)
	assert.Equal(t, "abc123", token)
}
