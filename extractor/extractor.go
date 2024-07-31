package extractor

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var ErrNoTokenInContext = errors.New("no token present in context")
var ErrInvalidTokenType = errors.New("invalid token type")

type Extractor interface {
	Extract(ctx context.Context) (string, error)
	ExtractRequest(http *http.Request) (string, error)
}

type HeaderExtractor []string

func (h HeaderExtractor) Extract(ctx context.Context) (string, error) {
	for _, header := range h {
		if token := ctx.Value(header); token != nil {
			if tokenStr, ok := token.(string); ok {
				return tokenStr, nil
			}
		}
	}
	return "", ErrNoTokenInContext
}

func (h HeaderExtractor) ExtractRequest(r *http.Request) (string, error) {
	for _, header := range h {
		token := r.Header.Get(header)
		if token != "" {
			return token, nil
		}
	}
	return "", ErrNoTokenInContext
}

type BearerExtractor struct{}

func (e BearerExtractor) Extract(ctx context.Context) (string, error) {
	token := ctx.Value("Authorization")
	token_header, ok := token.(string)
	if !ok {
		return "", ErrInvalidTokenType
	}
	if token_header == "" || !strings.HasPrefix(strings.ToLower(token_header), "bearer ") {
		return "", ErrNoTokenInContext
	}
	return token_header[7:], nil
}

func (e BearerExtractor) ExtractRequest(r *http.Request) (string, error) {
	token_header := r.Header.Get("Authorization")
	if token_header == "" || !strings.HasPrefix(strings.ToLower(token_header), "bearer ") {
		return "", ErrNoTokenInContext
	}
	return token_header[7:], nil
}

type AuthorizationExtractor struct{}

func (e AuthorizationExtractor) Extract(ctx context.Context) (string, error) {
	token := ctx.Value("Authorization")
	token_header, ok := token.(string)
	if !ok {
		return "", ErrInvalidTokenType
	}
	if token_header == "" {
		return "", ErrNoTokenInContext
	}
	return token_header, nil
}

func (e AuthorizationExtractor) ExtractRequest(r *http.Request) (string, error) {
	token_header := r.Header.Get("Authorization")
	if token_header == "" {
		return "", ErrNoTokenInContext
	}
	return token_header, nil
}

type TokenExtractor struct{}

func (e TokenExtractor) Extract(ctx context.Context) (string, error) {
	token := ctx.Value("Token")
	token_header, ok := token.(string)
	if !ok {
		return "", ErrInvalidTokenType
	}
	if token_header == "" {
		return "", ErrNoTokenInContext
	}
	return token_header, nil
}

func (e TokenExtractor) ExtractRequest(r *http.Request) (string, error) {
	token_header := r.Header.Get("Token")
	if token_header == "" {
		return "", ErrNoTokenInContext
	}
	return token_header, nil
}
