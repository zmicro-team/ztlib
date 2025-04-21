package extractor

import (
	"context"
	"net/http"
	"net/url"
)

var _ Extractor = (*AuthorizationExtractor)(nil)

type AuthorizationExtractor struct{}

func NewAuthorizationExtractor() AuthorizationExtractor {
	return AuthorizationExtractor{}
}

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

func (e AuthorizationExtractor) ExtractQuery(query url.Values) (string, error) {
	tokenHeader := query.Get("Authorization")
	if tokenHeader == "" {
		return "", ErrNoTokenInContext
	}
	return tokenHeader, nil
}
