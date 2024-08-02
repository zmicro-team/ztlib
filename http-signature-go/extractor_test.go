package httpsign

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

var errTestExtractor = errors.New("test")

type testExtractor string

func (h testExtractor) Extract(*http.Request) (string, Scheme, error) {
	return "", SchemeUnspecified, errTestExtractor
}

func Test_Extractor(t *testing.T) {
	testSignatureValue := "aaa"
	testAuthorizationSignatureValue := headerValueAuthorizationInitPrefix + "aaa"

	tests := []struct {
		name          string
		extractor     Extractor
		headers       map[string]string
		wantSignature string
		wantScheme    Scheme
		wantErr       error
	}{
		{
			name:          "header - Signature",
			extractor:     NewSignatureExtractor(HeaderSignature),
			headers:       map[string]string{HeaderSignature: testSignatureValue},
			wantSignature: testSignatureValue,
			wantScheme:    SchemeSignature,
			wantErr:       nil,
		},
		{
			name:          "header - Authorization",
			extractor:     NewAuthorizationSignatureExtractor(HeaderAuthorization),
			headers:       map[string]string{HeaderAuthorization: testAuthorizationSignatureValue},
			wantSignature: testSignatureValue,
			wantScheme:    SchemeAuthentication,
			wantErr:       nil,
		},
		{
			name: "header - multiple extractors first match",
			extractor: MultiExtractor{
				NewSignatureExtractor(HeaderSignature),
				NewAuthorizationSignatureExtractor(HeaderAuthorization),
			},
			headers:       map[string]string{HeaderSignature: testSignatureValue},
			wantSignature: testSignatureValue,
			wantScheme:    SchemeSignature,
			wantErr:       nil,
		},
		{
			name: "header - multiple extractors second match",
			extractor: NewMultiExtractor(
				NewSignatureExtractor(HeaderSignature),
				NewAuthorizationSignatureExtractor(HeaderAuthorization),
			),
			headers:       map[string]string{HeaderAuthorization: testAuthorizationSignatureValue},
			wantSignature: testSignatureValue,
			wantScheme:    SchemeAuthentication,
			wantErr:       nil,
		},
		{
			name:          "header - Signature miss",
			extractor:     NewSignatureExtractor(HeaderSignature),
			headers:       map[string]string{"miss": testSignatureValue},
			wantSignature: "",
			wantScheme:    SchemeUnspecified,
			wantErr:       ErrNoSignatureInRequest,
		},
		{
			name:          "header - Authorization miss",
			extractor:     NewAuthorizationSignatureExtractor(HeaderAuthorization),
			headers:       map[string]string{"miss": testAuthorizationSignatureValue},
			wantSignature: "",
			wantScheme:    SchemeUnspecified,
			wantErr:       ErrNoSignatureInRequest,
		},
		{
			name: "header - multiple extractors miss",
			extractor: NewMultiExtractor(
				NewSignatureExtractor(HeaderSignature),
				NewAuthorizationSignatureExtractor(HeaderAuthorization),
			),
			headers:       map[string]string{"miss": testSignatureValue},
			wantSignature: "",
			wantScheme:    SchemeUnspecified,
			wantErr:       ErrNoSignatureInRequest,
		},
		{
			name: "header - multiple extractors not ErrNoSignatureInRequest error",
			extractor: NewMultiExtractor(
				testExtractor("xx"),
				NewSignatureExtractor(HeaderSignature),
			),
			headers:       map[string]string{"miss": testSignatureValue},
			wantSignature: "",
			wantScheme:    SchemeUnspecified,
			wantErr:       errTestExtractor,
		},
	}

	for _, tt := range tests {
		r := makeTestRequest("GET", "/", tt.headers)
		token, scheme, err := tt.extractor.Extract(r)
		require.Equal(t, tt.wantErr, err)
		require.Equal(t, tt.wantSignature, token)
		require.Equal(t, tt.wantScheme, scheme)
	}
}

func makeTestRequest(method, path string, headers map[string]string) *http.Request {
	r, _ := http.NewRequestWithContext(context.Background(), method, path, nil)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}
