package extractor

import (
	"context"
	"net/http"
	"net/url"
)

type TokenExtractor struct{}

var _ Extractor = (*TokenExtractor)(nil)

func NewTokenExtractor() TokenExtractor {
	return TokenExtractor{}
}

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

func (e TokenExtractor) ExtractQuery(query url.Values) (string, error) {
	tokenHeader := query.Get("Token")
	if tokenHeader == "" {
		return "", ErrNoTokenInContext
	}
	return tokenHeader, nil
}
