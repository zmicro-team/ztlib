package signature

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	httpsign "github.com/zmicro-team/ztlib/http-signature-go"
)

type ctxKeyIdKey struct{}

func WithKeyId(ctx context.Context, keyId httpsign.KeyId) context.Context {
	return context.WithValue(ctx, ctxKeyIdKey{}, keyId)
}

func FromKeyId(ctx context.Context) (v httpsign.KeyId, ok bool) {
	v, ok = ctx.Value(ctxKeyIdKey{}).(httpsign.KeyId)
	return
}

func MustFromKeyId(ctx context.Context) httpsign.KeyId {
	v, ok := ctx.Value(ctxKeyIdKey{}).(httpsign.KeyId)
	if !ok {
		panic("signature: must be set keyId in context")
	}
	return v
}

// Authenticator is the gin authenticator middleware.
type Authenticator struct {
	Parser      *httpsign.Parser
	ErrFallback func(c *gin.Context, statusCode int, err error)
}

// Authenticated returns a gin middleware which permits given permissions in parameter.
func (a *Authenticator) Authenticated() gin.HandlerFunc {
	if a.Parser == nil {
		panic("http sign parse must be initialized.")
	}
	if a.ErrFallback == nil {
		a.ErrFallback = func(c *gin.Context, statusCode int, err error) {
			c.String(statusCode, err.Error())
		}
	}
	return func(c *gin.Context) {
		parameter, err := a.Parser.ParseFromRequest(c.Request)
		if err != nil {
			c.Abort()
			a.ErrFallback(c, http.StatusBadRequest, err)
			return
		}
		err = a.Parser.Verify(c.Request, parameter)
		if err != nil {
			statusCode := http.StatusBadRequest
			if parameter.Scheme != httpsign.SchemeUnspecified &&
				errors.Is(err, httpsign.ErrSignatureInvalid) {
				if parameter.Scheme == httpsign.SchemeSignature {
					statusCode = http.StatusForbidden
				} else {
					statusCode = http.StatusUnauthorized
				}
			}
			c.Abort()
			a.ErrFallback(c, statusCode, err)
			return
		}
		c.Request = c.Request.WithContext(WithKeyId(c.Request.Context(), parameter.KeyId))
		c.Next()
	}
}