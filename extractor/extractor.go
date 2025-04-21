package extractor

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

var ErrNoTokenInContext = errors.New("no token present in context")
var ErrInvalidTokenType = errors.New("invalid token type")

type Extractor interface {
	Extract(ctx context.Context) (string, error)
	ExtractRequest(http *http.Request) (string, error)
	ExtractHeader(header http.Header) (string, error)
	ExtractQuery(query url.Values) (string, error)
}
