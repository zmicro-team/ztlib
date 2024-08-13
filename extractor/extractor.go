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
	ExtractHeader(header http.Header) (string, error)
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

func (h HeaderExtractor) ExtractHeader(httpHeader http.Header) (string, error) {
	for _, header := range h {
		token := httpHeader.Get(header)
		if token != "" {
			return token, nil
		}
	}
	return "", ErrNoTokenInContext
}

type BearerExtractor struct{}

func (e BearerExtractor) Extract(ctx context.Context) (string, error) {
	token := ctx.Value("Authorization")
	tokenHeader, ok := token.(string)
	if !ok {
		return "", ErrInvalidTokenType
	}
	if tokenHeader == "" || !strings.HasPrefix(strings.ToLower(tokenHeader), "bearer ") {
		return "", ErrNoTokenInContext
	}
	return tokenHeader[7:], nil
}

func (e BearerExtractor) ExtractRequest(r *http.Request) (string, error) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" || !strings.HasPrefix(strings.ToLower(tokenHeader), "bearer ") {
		return "", ErrNoTokenInContext
	}
	return tokenHeader[7:], nil
}

func (e BearerExtractor) ExtractHeader(header http.Header) (string, error) {
	tokenHeader := header.Get("Authorization")
	if tokenHeader == "" || !strings.HasPrefix(strings.ToLower(tokenHeader), "bearer ") {
		return "", ErrNoTokenInContext
	}
	return tokenHeader[7:], nil
}

type AuthorizationExtractor struct{}

func (e AuthorizationExtractor) Extract(ctx context.Context) (string, error) {
	token := ctx.Value("Authorization")
	tokenHeader, ok := token.(string)
	if !ok {
		return "", ErrInvalidTokenType
	}
	if tokenHeader == "" {
		return "", ErrNoTokenInContext
	}
	return tokenHeader, nil
}

func (e AuthorizationExtractor) ExtractRequest(r *http.Request) (string, error) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" {
		return "", ErrNoTokenInContext
	}
	return tokenHeader, nil
}

func (e AuthorizationExtractor) ExtractHeader(header http.Header) (string, error) {
	tokenHeader := header.Get("Authorization")
	if tokenHeader == "" {
		return "", ErrNoTokenInContext
	}
	return tokenHeader, nil
}

type TokenExtractor struct{}

func (e TokenExtractor) Extract(ctx context.Context) (string, error) {
	token := ctx.Value("Token")
	tokenHeader, ok := token.(string)
	if !ok {
		return "", ErrInvalidTokenType
	}
	if tokenHeader == "" {
		return "", ErrNoTokenInContext
	}
	return tokenHeader, nil
}

func (e TokenExtractor) ExtractRequest(r *http.Request) (string, error) {
	tokenHeader := r.Header.Get("Token")
	if tokenHeader == "" {
		return "", ErrNoTokenInContext
	}
	return tokenHeader, nil
}

func (e TokenExtractor) ExtractHeader(header http.Header) (string, error) {
	tokenHeader := header.Get("Token")
	if tokenHeader == "" {
		return "", ErrNoTokenInContext
	}
	return tokenHeader, nil
}
