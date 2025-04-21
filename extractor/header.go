package extractor

import (
	"context"
	"net/http"
	"net/url"
)

var _ Extractor = (*HeaderExtractor)(nil)

type HeaderExtractor []string

func NewHeaderExtractor(header ...string) HeaderExtractor {
	return header
}

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

func (h HeaderExtractor) ExtractQuery(query url.Values) (string, error) {
	for _, header := range h {
		token := query.Get(header)
		if token != "" {
			return token, nil
		}
	}
	return "", ErrNoTokenInContext
}
