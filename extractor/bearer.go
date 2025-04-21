package extractor

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

var _ Extractor = (*BearerExtractor)(nil)

type BearerExtractor struct{}

func NewBearerExtractor() BearerExtractor {
	return BearerExtractor{}
}

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

func (e BearerExtractor) ExtractQuery(query url.Values) (string, error) {
	tokenHeader := query.Get("Authorization")
	if tokenHeader == "" || !strings.HasPrefix(strings.ToLower(tokenHeader), "bearer ") {
		return "", ErrNoTokenInContext
	}
	return tokenHeader[7:], nil
}
